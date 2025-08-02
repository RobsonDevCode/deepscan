package scannermodels

type Project struct {
	ServiceName        string
	Name               string
	Ecosystem          string
	PackagesAndVersion map[string]string
	Framework          string
	Frameworks         string
}
