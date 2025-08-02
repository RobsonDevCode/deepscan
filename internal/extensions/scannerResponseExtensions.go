package extensions

import (
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
)

func FlatternPackages(scannedProjects []models.ScannerResponse) []models.ScannedPackage {
	var scannedPackages []models.ScannedPackage

	for _, scannedProject := range scannedProjects {
		for _, pkg := range scannedProject.Packages {
			pkg.ServiceName = scannedProject.ServiceName
			pkg.ProjectName = scannedProject.Name
			scannedPackages = append(scannedPackages, pkg)
		}
	}
	return scannedPackages
}
