package githubreposmodels

type GithubRepository struct {
	Name     string `json:"full_name"`
	Private  bool   `json:"private"`
	CloneUrl string `json:"clone_url"`
}
