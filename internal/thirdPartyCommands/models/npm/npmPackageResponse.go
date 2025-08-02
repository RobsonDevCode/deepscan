package npmmodels

type NpmPackageResponse struct {
	Version     string                `json:"version"`
	ServiceName string                `json:"name"`
	NpmPackage  map[string]NpmPackage `json:"dependencies"`
}
