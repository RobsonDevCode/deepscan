package scannermapper

import (
	"fmt"

	ecosystemconstants "github.com/RobsonDevCode/deepscan/internal/scanner/constants/ecosystem"
	scannermodels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
	npmmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models/npm"
)

func MapNpmResultToProject(response npmmodels.NpmPackageResponse) scannermodels.Project {
	return scannermodels.Project{
		ServiceName:        response.ServiceName,
		Name:               response.ServiceName,
		Ecosystem:          ecosystemconstants.Npm,
		PackagesAndVersion: MapPackageAndVersion(response.NpmPackage),
	}
}

func MapPackageAndVersion(packages map[string]npmmodels.NpmPackage) map[string]string {
	result := make(map[string]string)
	var flattern func(map[string]npmmodels.NpmPackage)
	flattern = func(pkgs map[string]npmmodels.NpmPackage) {
		for pkg, packageInfo := range pkgs {
			if _, exists := result[pkg]; !exists {
				result[pkg] = packageInfo.Version
				fmt.Printf("Package: '%s' Version: '%s'\n", pkg, packageInfo.Version)
			}

			//Recursive call on dependency tree as packages can have multiple nodes of dependencies
			if len(packageInfo.Dependencies) > 0 {
				flattern(packageInfo.Dependencies)
			}
		}
	}

	flattern(packages)
	return result
}
