package scannerService

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/RobsonDevCode/deepscan/internal/clients"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	"github.com/RobsonDevCode/deepscan/internal/extensions"
	scannerconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants"
	ecosystemconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants/ecosystem"
	riskscoreconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants/riskScore"
	scannermodels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
	projectreaderservice "github.com/RobsonDevCode/deepscan/internal/services/projectReaderService"
	"golang.org/x/sync/errgroup"
)

// Githubs max request size
const batchSize = 100

type ScannerService interface {
	ScanProject(root string, ctx context.Context) ([]models.ScannerResponse, error)
	ScanProjects(ctx context.Context) (models.ScanAllResponse, error)
}

type Scanner struct {
	client        clients.GithubClientService
	packageReader projectreaderservice.PackageReaderService
}

func NewScanner(client clients.GithubClientService,
	packageReader projectreaderservice.PackageReaderService) *Scanner {
	return &Scanner{
		client:        client,
		packageReader: packageReader,
	}
}

func (s *Scanner) ScanProjects(ctx context.Context) (models.ScanAllResponse, error) {
	defer os.RemoveAll(scannerconstants.TempDirctory)
	projectFiles, err := s.GetFilesToScan(scannerconstants.TempDirctory, ctx)
	if err != nil {
		return models.ScanAllResponse{}, err
	}

	if projectFiles == nil {
		return models.ScanAllResponse{}, fmt.Errorf("project files are empty")
	}

	//we get 429 and 403 from the github api so we have to put a limiter on the concurrent channels
	maxConcurrentChans := 2
	scans := make(chan scannermodels.ConcurrentScanResult, maxConcurrentChans)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, projectFile := range projectFiles {
		wg.Add(1)
		go func(pf scannermodels.Project) {
			var err error
			defer wg.Done()

			packageInfo, err := s.validateAndScan(pf, &mu, ctx)
			if err != nil {
				scans <- scannermodels.ConcurrentScanResult{
					Project:     nil,
					Err:         err,
					ServiceName: pf.ServiceName,
					ProjectName: pf.Name,
				}
			}

			var framework string
			if projectFile.Framework == "" {
				framework = pf.Frameworks
			} else {
				framework = pf.Framework
			}

			setRiskScore(packageInfo)

			scannerResponse := &models.ScannerResponse{
				Packages:    packageInfo,
				Framework:   framework,
				Name:        pf.Name,
				ServiceName: pf.ServiceName,
			}

			scans <- scannermodels.ConcurrentScanResult{
				Project:     scannerResponse,
				Err:         err,
				ServiceName: pf.ServiceName,
				ProjectName: pf.Name,
			}

		}(projectFile)
	}

	go func() {
		wg.Wait()
		close(scans)
	}()

	result := extensions.MapScanAllResponse(scans)
	return result, nil
}

func (s *Scanner) ScanProject(root string, ctx context.Context) ([]models.ScannerResponse, error) {
	defer scannerCleanUp()
	projectFiles, err := s.GetFilesToScan(root, ctx)
	if err != nil {
		return nil, err
	}

	var result []models.ScannerResponse
	group, gCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	for _, projectFile := range projectFiles {
		group.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()

			default:
				//only need to check frameworks for cs projects
				packageInfo, err := s.validateAndScan(projectFile, &mu, gCtx)
				if err != nil {
					return err
				}

				var framework string
				if projectFile.Framework == "" {
					framework = projectFile.Frameworks
				} else {
					framework = projectFile.Framework
				}

				scannerResponse := &models.ScannerResponse{
					Packages:    packageInfo,
					Framework:   framework,
					Name:        projectFile.Name,
					ServiceName: projectFile.ServiceName,
				}

				mu.Lock()
				result = append(result, *scannerResponse)
				mu.Unlock()

				return nil
			}
		})
	}

	if concurrentErr := group.Wait(); concurrentErr != nil {
		return nil, concurrentErr
	}

	return result, nil
}

func (s *Scanner) GetFilesToScan(root string, ctx context.Context) ([]scannermodels.Project, error) {
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	var projects []scannermodels.Project

	walkErr := filepath.WalkDir(root, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking dir: %w", err)
		}

		supported, projectType := s.packageReader.IsSupportedProject(path)
		if !supported {
			//skip dir
			return nil
		}

		if !dir.IsDir() && supported {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return fmt.Errorf("task has been cancelled, %w", ctx.Err())

				default:
					project, err := s.packageReader.ReadProject(projectType, path, ctx)
					if err != nil {
						return err
					}

					mu.Lock()
					projects = append(projects, project)
					mu.Unlock()

					return nil
				}
			})
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *Scanner) validateAndScan(projectFile scannermodels.Project, mu *sync.Mutex, ctx context.Context) ([]models.ScannedPackage, error) {
	//only need to check frameworks for cs projects
	if (projectFile.Framework == "" && projectFile.Frameworks == "") &&
		projectFile.Ecosystem == ecosystemconstants.Nuget {
		return nil, fmt.Errorf("\ncouldnt get framework for: %s", projectFile.Name)
	}

	if len(projectFile.PackagesAndVersion) == 0 {
		fmt.Printf("\nFile has no packages, skipping")
		return nil, nil
	}

	var packageInfo []models.ScannedPackage
	packagesLength := len(projectFile.PackagesAndVersion)

	if packagesLength >= batchSize {
		responses, err := s.sendBatchedPackages(projectFile, mu, ctx)
		if err != nil {
			return nil, fmt.Errorf("error handling batched packages %s", err)
		}
		packageInfo = responses
	} else {
		response, err := s.client.GetPackagesInfo(projectFile.Ecosystem, projectFile.PackagesAndVersion, ctx)
		if err != nil {
			return nil, fmt.Errorf("\nerror, getting packages from client: %w", err)
		}
		packageInfo = response
	}

	return packageInfo, nil
}

func (s *Scanner) sendBatchedPackages(projectFile scannermodels.Project, mu *sync.Mutex, ctx context.Context) ([]models.ScannedPackage, error) {
	batch := make(map[string]string)
	itemCount := 0

	var result []models.ScannedPackage

	for packageName, version := range projectFile.PackagesAndVersion {
		batch[packageName] = version
		itemCount++

		if len(batch) == batchSize {
			// Process full batch
			response, err := s.client.GetPackagesInfo(projectFile.Ecosystem, batch, ctx)
			if err != nil {
				return nil, fmt.Errorf("error getting packages from client: %w", err)
			}

			mu.Lock()
			result = append(result, response...)
			mu.Unlock()

			batch = make(map[string]string)
			itemCount = 0
		}
	}

	// Handle remaining items
	if len(batch) > 0 {
		response, err := s.client.GetPackagesInfo(projectFile.Ecosystem, batch, ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting packages from client: %w", err)
		}

		mu.Lock()
		result = append(result, response...)
		fmt.Printf("\nLength of result: %d\n", len(result))
		mu.Unlock()
	}

	return result, nil
}

func scannerCleanUp() error {
	if _, err := os.Stat(scannerconstants.TempDirctory); err == nil {
		os.RemoveAll(scannerconstants.TempDirctory)
		return err
	}

	return nil
}

func setRiskScore(packages []models.ScannedPackage) {

	for i := range packages {
		switch packages[i].Severity {
		case "critical":
			packages[i].RiskScore = riskscoreconstants.Critical
		case "high":
			packages[i].RiskScore = riskscoreconstants.High
		case "medium":
			packages[i].RiskScore = riskscoreconstants.Medium
		case "low":
			packages[i].RiskScore = riskscoreconstants.Low
		}
	}
}
