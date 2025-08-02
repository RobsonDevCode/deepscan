package models

type ScannerResponse struct {
	Packages       []ScannedPackage
	Framework      string
	Name           string
	ServiceName    string
	CurrentVersion string
}
