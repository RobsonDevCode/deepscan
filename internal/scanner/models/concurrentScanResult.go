package scannermodels

import "github.com/RobsonDevCode/deepscan/internal/clients/models"

type ConcurrentScanResult struct {
	Project     *models.ScannerResponse
	Err         error
	ServiceName string
	ProjectName string
	PackageName string
}
