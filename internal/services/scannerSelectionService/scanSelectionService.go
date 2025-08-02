package scannerselectionservice

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	tablewriterservice "github.com/RobsonDevCode/deepscan/internal/cmdLineWriters/tablewriter"
	"github.com/RobsonDevCode/deepscan/internal/extensions"
	repositoryreaderservice "github.com/RobsonDevCode/deepscan/internal/services/repositoryReaderService"
	scanfileservice "github.com/RobsonDevCode/deepscan/internal/services/scanFileService"
	scansshservice "github.com/RobsonDevCode/deepscan/internal/services/scanShhService"
	setupservice "github.com/RobsonDevCode/deepscan/internal/services/setupService"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ScannerSelectionService interface {
	Scan(cmd *cobra.Command, ctx context.Context) ([]models.ScannedPackage, error)
	ScanAll(ctx context.Context) ([]models.ScannedPackage, error)
}

type ScanSelection struct {
	sshService             scansshservice.ScanSSHService
	fileService            scanfileservice.ScanFileService
	repositoryReaderFacade repositoryreaderservice.RepositoryReaderFacade
}

func NewScanSelection(sshService scansshservice.ScanSSHService,
	fileService scanfileservice.ScanFileService,
	repositoryReader repositoryreaderservice.RepositoryReaderFacade) ScanSelection {
	return ScanSelection{
		sshService:             sshService,
		fileService:            fileService,
		repositoryReaderFacade: repositoryReader,
	}
}

const (
	DirFlag = "dir"
	SSHFlag = "ssh"
)

func (s *ScanSelection) Scan(cmd *cobra.Command, ctx context.Context) ([]models.ScannedPackage, error) {
	filePath, _ := cmd.Flags().GetString(DirFlag)
	sshUrl, _ := cmd.Flags().GetString(SSHFlag)

	var scannedProjects []models.ScannerResponse
	if filePath != "" {
		scannerResponse, err := s.fileService.ScanProjectFile(filePath, ctx)
		if err != nil {
			return nil, err
		}

		scannedProjects = scannerResponse
	} else if sshUrl != "" {
		project, err := s.sshService.Scan(sshUrl, ctx)
		if err != nil {
			return nil, err
		}

		scannedProjects = project
	} else {
		selectedSshUrl, err := s.SelectFromAllProjects()
		if err != nil {
			return nil, err
		}

		scannerResponse, err := s.sshService.Scan(*selectedSshUrl, ctx)
		if err != nil {
			return nil, err
		}

		scannedProjects = scannerResponse
	}

	tablewriterservice.DisplayInfomationTable(scannedProjects)
	scannedPackages := extensions.FlatternPackages(scannedProjects)
	tablewriterservice.DisplayPackagesTable(scannedPackages)

	return scannedPackages, nil
}

func (s *ScanSelection) ScanAll(ctx context.Context) ([]models.ScannedPackage, error) {
	fmt.Print("Starting Scan...\n")
	scanAllResponse, err := s.sshService.CloneAndScanAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s", color.RedString(err.Error()))
	}

	scannedPackages := extensions.FlatternPackages(scanAllResponse.SuccessfullyScannedProjects)

	tablewriterservice.DisplayPackagesTable(scannedPackages)
	tablewriterservice.DisplayFailedScanTable(scanAllResponse.FailedProjects)

	return scannedPackages, nil
}

func (s *ScanSelection) SelectFromAllProjects() (*string, error) {
	fmt.Print("Loading projects...")

	userSettings, err := setupservice.GetUserSettings()
	if err != nil {
		return nil, err
	}

	projects, err := s.repositoryReaderFacade.GetRepos(*userSettings)
	if err != nil {
		return nil, fmt.Errorf("error getting current projects, %v", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projests found")
	}

	var options []string
	for _, project := range projects {
		options = append(options, project.Name)
	}

	prompt := &survey.Select{
		Message: "Select a project to scan:",
		Options: options,
	}

	var selectedIndex int
	err = survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		fmt.Print("selection cancelled")
		return nil, fmt.Errorf("survey error: %w", err)
	}

	return &projects[selectedIndex].SSHUrl, nil
}
