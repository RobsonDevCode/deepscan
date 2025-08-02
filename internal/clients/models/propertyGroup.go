package models

type PropertyGroup struct {
	FrameworkVersion string `xml:"TargetFramework"`
	LangVersion      int16  `xml:"LangVersion"`
}
