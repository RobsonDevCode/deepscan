package packagereaderservice

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	projecttypessupported "github.com/RobsonDevCode/deepscan/internal/constants/projectTypesSupported"
	scannermodels "github.com/RobsonDevCode/deepscan/internal/scanner/models"
	npmmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models/npm"
	npmcommands "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/npmCommands"
	"golang.org/x/net/context"
)

type PackageReaderService interface {
	ReadCsProject(path *string, ctx context.Context) (scannermodels.CsProject, error)
	ReadFrontEndProject(path *string, ctx context.Context) (npmmodels.NpmPackageResponse, error)
	GetProjectType(root string, ctx context.Context) (*string, *string, error)
}

type PackageReader struct{}

func NewPackageReader() *PackageReader {
	return &PackageReader{}
}

func (r *PackageReader) GetProjectType(root string, ctx context.Context) (*string, *string, error) {
	var projectType string
	var pathFound string

	err := filepath.WalkDir(root, func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking dir: %w", err)
		}

		if !dir.IsDir() && strings.HasSuffix(path, projecttypessupported.Npm) {
			projectType = projecttypessupported.Npm
			// we save the file path if its a npm project for quicker auditing
			pathFound = filepath.Dir(path)
			return nil
		}

		if !dir.IsDir() && strings.HasSuffix(path, projecttypessupported.Dotnet) {
			projectType = projecttypessupported.Dotnet
			return nil
		}

		return fmt.Errorf("error project type not supported")
	})
	if err != nil {
		return nil, nil, err
	}

	return &projectType, &pathFound, nil
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
