package scanfileservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	scannerService "github.com/RobsonDevCode/deepscan/internal/scanner"
	projectreaderservice "github.com/RobsonDevCode/deepscan/internal/services/projectReaderService"
	"github.com/fatih/color"
)

type ScanFileService interface {
	ScanProjectFile(filepath string, ctx context.Context) ([]models.ScannerResponse, error)
}

type FileProcessor struct {
	scanner       scannerService.ScannerService
	packageReader projectreaderservice.PackageReaderService
}

func NewFileScannerService(scanner scannerService.ScannerService, packageReader projectreaderservice.PackageReaderService) *FileProcessor {
	return &FileProcessor{
		scanner:       scanner,
		packageReader: packageReader,
	}
}

func (f *FileProcessor) ScanProjectFile(filePath string, ctx context.Context) ([]models.ScannerResponse, error) {
	parts := strings.Split(filePath, "\\")
	selectedProject := &parts[(len(parts) - 1)]

	fmt.Printf("Selected Project: %s \n", color.CyanString("%s", *selectedProject))

	scannedProject, err := f.scanner.ScanProject(filePath, ctx)
	if err != nil {
		return nil, err
	}

	if scannedProject == nil {
		return nil, fmt.Errorf("error, scanned project returned nil")
	}

	if len(scannedProject) == 0 {
		fmt.Printf("No vunrabilities on project %v", selectedProject)
	}

	return scannedProject, nil
}
