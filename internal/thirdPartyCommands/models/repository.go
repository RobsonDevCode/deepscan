package cmdmodels

type Repository struct {
	SSHUrl     string `json:"sshUrl"`
	Name       string `json:"name"`
	IsDisabled bool   `json:"isDisabled"`
}
