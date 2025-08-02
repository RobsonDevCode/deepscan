package scannermodels

type FrontEndProject struct {
	ProjectName string                     `json:"name,omitempty"`
	Packages    map[string]FrontEndPackage `json:"packages"`
}
