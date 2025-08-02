package scannermodels

import "github.com/RobsonDevCode/deepscan/internal/clients/models"

type CsProject struct {
	Framework         string                    `xml:"PropertyGroup>TargetFramework"`
	Frameworks        string                    `xml:"PropertyGroup>TargetFrameworks"`
	PackageReferences []models.PackageReference `xml:"ItemGroup>PackageReference"`
	Name              string                    `xml:"-"`
	ServiceName       string                    `xml:"-"`
}
