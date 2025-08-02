package models

type FailedProjectScan struct {
	Error       error
	ServiceName string
	ProjectName string
	PackageName string
}
