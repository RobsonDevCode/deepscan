package scanfileservice

import (
	"context"
	"fmt"
	"strings"

	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	scannerService "github.com/RobsonDevCode/deepscan/internal/scanner"
	packagereaderservice "github.com/RobsonDevCode/deepscan/internal/services/packageReaderService"
	"github.com/fatih/color"
)

type ScanFileService interface {
	ScanProjectFile(filepath string, ctx context.Context) ([]models.ScannerResponse, error)
}

type FileProcessor struct {
	scanner       scannerService.ScannerService
	packageReader packagereaderservice.PackageReaderService
}

func NewFileScannerService(scanner scannerService.ScannerService, packageReader packagereaderservice.PackageReaderService) *FileProcessor {
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
