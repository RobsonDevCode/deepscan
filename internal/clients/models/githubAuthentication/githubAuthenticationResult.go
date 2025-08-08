package authenticaionmodels

type GithubAuthenticationResult struct {
	AccessToken *GithubAccessToken
	AuthError   *AuthenticationError
}
