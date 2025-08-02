package scannermodels

type FrontEndPackage struct {
	Name         string            `json:"name,omitempty"`
	Version      string            `json:"version,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"` // packageName -key version - value
}
