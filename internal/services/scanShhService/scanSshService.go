package scansshservice

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	scannerService "github.com/RobsonDevCode/deepscan/internal/scanner"
	scannerconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants"
	repositoryreaderservice "github.com/RobsonDevCode/deepscan/internal/services/repositoryReaderService"
	setupservice "github.com/RobsonDevCode/deepscan/internal/services/setupService"
	githubcommands "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/githubCommands"
	"github.com/fatih/color"
)

type ScanSSHService interface {
	Scan(sshUrl string, ctx context.Context) ([]models.ScannerResponse, error)
	CloneAndScanAll(ctx context.Context) (models.ScanAllResponse, error)
}

type SShProcessor struct {
	scanner                scannerService.ScannerService
	repositoryReaderFacade repositoryreaderservice.RepositoryReaderFacade
}

func NewSshProcessor(scanner scannerService.ScannerService, repositoryReader repositoryreaderservice.RepositoryReaderFacade) *SShProcessor {
	return &SShProcessor{
		scanner:                scanner,
		repositoryReaderFacade: repositoryReader,
	}
}

func (s *SShProcessor) Scan(sshUrl string, ctx context.Context) ([]models.ScannerResponse, error) {
	parts := strings.Split(sshUrl, "/")
	selectedProject := parts[(len(parts) - 1)]
	fmt.Printf("Selected Project: %s \n", color.CyanString("%s", selectedProject))

	err := githubcommands.CloneRepository(sshUrl, ctx)
	if err != nil {
		return nil, fmt.Errorf("error cloning %s error: %w", selectedProject, err)
	}

	scannedProject, err := s.scanner.ScanProject(scannerconstants.TempDirctory, ctx)
	if err != nil {
		return nil, err
	}

	return scannedProject, nil
}

func (s *SShProcessor) CloneAndScanAll(ctx context.Context) (models.ScanAllResponse, error) {
	userSettings, err := setupservice.GetUserSettings()
	if err != nil {
		return models.ScanAllResponse{}, err
	}

	repos, err := s.repositoryReaderFacade.GetRepos(*userSettings, ctx)
	if err != nil {
		return models.ScanAllResponse{}, err
	}

	if repos == nil {
		return models.ScanAllResponse{}, fmt.Errorf("repositories return nil")
	}

	var sshUrls []string
	for _, repo := range repos {
		sshUrls = append(sshUrls, repo.SSHUrl)
	}
	if len(sshUrls) == 0 {
		return models.ScanAllResponse{}, fmt.Errorf("error: sshUrls cannot be nil or empty when trying to clone and scan")
	}

	if err := githubcommands.CloneAll(sshUrls, ctx); err != nil {
		deleteErr := os.RemoveAll(scannerconstants.TempDirctory)
		if deleteErr != nil {
			return models.ScanAllResponse{}, deleteErr
		}

		return models.ScanAllResponse{}, fmt.Errorf("error cloning all repos: %w", err)
	}

	scannedProjects, err := s.scanner.ScanProjects(ctx)
	if err != nil {
		return models.ScanAllResponse{}, err
	}
	return scannedProjects, nil
}
