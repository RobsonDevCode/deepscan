package scannermapper

import (
	ecosystemconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants/ecosystem"
	scannermodels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
)

func MapCsProjToProject(csProject *scannermodels.CsProject) scannermodels.Project {
	return scannermodels.Project{
		ServiceName:        csProject.ServiceName,
		Name:               csProject.Name,
		Ecosystem:          ecosystemconstants.Nuget,
		PackagesAndVersion: CsProjSliceToMap(*csProject),
		Framework:          csProject.Framework,
		Frameworks:         csProject.Frameworks,
	}
}

func CsProjSliceToMap(csproj scannermodels.CsProject) map[string]string {
	result := make(map[string]string)
	for _, pkg := range csproj.PackageReferences {
		result[pkg.Name] = pkg.Version
	}
	return result
}
