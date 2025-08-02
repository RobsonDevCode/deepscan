package models

type ScanAllResponse struct {
	SuccessfullyScannedProjects []ScannerResponse
	FailedProjects              []FailedProjectScan
}
