package npmmodels

type NpmPackage struct {
	Version string `json:"version"`
	//Npm package-lock.json has nested dependencies on packages we also need to scan for
	Dependencies map[string]NpmPackage `json:"dependencies,omitempty"`
}
