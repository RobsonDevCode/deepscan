package models

type GithubAuthenticationResult struct {
	AccessToken *GithubAccessToken
	AuthError   *AuthenticationError
}
