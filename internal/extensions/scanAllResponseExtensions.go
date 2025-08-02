package extensions

import (
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	scannerModels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
)

func MapScanAllResponse(scans chan scannerModels.ConcurrentScanResult) models.ScanAllResponse {
	var scannedProjects []models.ScannerResponse
	var failed []models.FailedProjectScan

	for scan := range scans {
		if scan.Err != nil {
			failedScan := &models.FailedProjectScan{
				Error:       scan.Err,
				ServiceName: scan.ServiceName,
				ProjectName: scan.ProjectName,
				PackageName: scan.PackageName,
			}

			failed = append(failed, *failedScan)
		} else {
			scannedProjects = append(scannedProjects, *scan.Project)
			fmt.Printf("\nSuccessfully processed %s", scan.ServiceName)
		}

	}

	result := models.ScanAllResponse{
		SuccessfullyScannedProjects: scannedProjects,
		FailedProjects:              failed,
	}

	return result
}
