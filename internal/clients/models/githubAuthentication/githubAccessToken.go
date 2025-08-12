package authenticaionmodels

type GithubAccessToken struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	Scope     string `json:"scope"`
}
