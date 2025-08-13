package projectreaderservice

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RobsonDevCode/deepscan/internal/constants/supportedprojects"
	scannermapper "github.com/RobsonDevCode/deepscan/internal/scanner/mapping"
	scannermodels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
	npmmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models/npm"
	npmcommands "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/npmCommands"
	"golang.org/x/mod/modfile"
	"golang.org/x/net/context"
)

type PackageReaderService interface {
	ReadCsProject(path *string, ctx context.Context) (scannermodels.CsProject, error)
	ReadFrontEndProject(path *string, ctx context.Context) (npmmodels.NpmPackageResponse, error)
	ReadProject(projectType string, path string, ctx context.Context) (scannermodels.Project, error)
	IsSupportedProject(path string) (bool, string)
}

type PackageReader struct{}

func NewPackageReader() *PackageReader {
	return &PackageReader{}
}

func (r *PackageReader) IsSupportedProject(path string) (bool, string) {
	if strings.HasSuffix(path, ".csproj") {
		return true, supportedprojects.CsProject
	}

	if strings.HasSuffix(path, "package-lock.json") {
		return true, supportedprojects.Npm
	}

	if strings.HasSuffix(path, ".mod") {
		return true, supportedprojects.GoProject
	}

	return false, ""
}

func (r *PackageReader) ReadProject(projectType string, path string, ctx context.Context) (scannermodels.Project, error) {
	switch projectType {
	case supportedprojects.CsProject:
		csProject, err := r.ReadCsProject(&path, ctx)
		if err != nil {
			return scannermodels.Project{}, fmt.Errorf("error reading C# project %w", err)
		}

		return scannermapper.MapCsProjToProject(&csProject), nil

	case supportedprojects.Npm:
		dirToScan := filepath.Dir(path)
		npmProject, err := r.ReadFrontEndProject(&dirToScan, ctx)
		if err != nil {
			return scannermodels.Project{}, fmt.Errorf("error reading Npm project  %w", err)
		}

		return scannermapper.MapNpmResultToProject(npmProject), nil
	case supportedprojects.GoProject:
		return scannermodels.Project{}, nil

	default:
		return scannermodels.Project{}, fmt.Errorf("unsupported type attempting to be scanned")
	}
}

func (r *PackageReader) ReadCsProject(path *string, ctx context.Context) (scannermodels.CsProject, error) {
	content, err := os.ReadFile(*path)
	if err != nil {
		return scannermodels.CsProject{}, fmt.Errorf("error reading csproj file %s error: %w", *path, err)
	}

	var project scannermodels.CsProject
	if err := xml.Unmarshal(content, &project); err != nil {
		return scannermodels.CsProject{}, fmt.Errorf("error unmarshalling xml file %s error: %w", *path, err)
	}

	parts := strings.Split(*path, "\\")
	project.Name = strings.TrimSuffix(parts[(len(parts)-1)], ".csproj")
	project.ServiceName = parts[1]

	return project, nil
}

func (r *PackageReader) ReadGoProject(path string) (scannermodels.GoProject, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return scannermodels.GoProject{}, fmt.Errorf("error reading go project file: %w", err)
	}

	modFile, err := modfile.Parse(path, data, nil)
	if err != nil {
		return scannermodels.GoProject{}, fmt.Errorf("error parsing mod file: %w", err)
	}

	serviceName := modFile.Module.Mod.Path
	return scannermodels.GoProject{
		ServiceName: serviceName,
		LangVersion: modFile.Go.Version,
		Packages:    modFile.Require,
	}, nil
}

func (r *PackageReader) ReadFrontEndProject(path *string, ctx context.Context) (npmmodels.NpmPackageResponse, error) {

	response, err := npmcommands.GetPackages(*path, ctx)
	if err != nil {
		return npmmodels.NpmPackageResponse{}, fmt.Errorf("error reading frontend %s error: %w", *path, err)
	}

	if len(response.NpmPackage) == 0 {
		return npmmodels.NpmPackageResponse{}, fmt.Errorf("\n error processing json packages are nil")
	}
	return response, nil
}
