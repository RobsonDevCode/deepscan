package configuration

import (
	authenticaionmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/githubAuthentication"
)

type UsersSettings struct {
	OrganizationUrl string                                 `json:"org_url"`
	Profile         string                                 `json:"profile"`
	Provider        string                                 `json:"provider"`
	AccessToken     *authenticaionmodels.GithubAccessToken `json:"access_token"`
}
