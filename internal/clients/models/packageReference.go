package models

type PackageReference struct {
	Name    string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}
