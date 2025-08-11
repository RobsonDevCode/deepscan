package gitubauthenticationservice

import (
	"context"
	"fmt"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	githubauthenticationclient "github.com/RobsonDevCode/deepscan/internal/clients/githubAuthenticationClient"
	authenticaionmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/githubAuthentication"
)

type GithubAuthenticator struct {
	githubAuthenticationClient githubauthenticationclient.GithubAuthenticationClientService
	cache                      *cache.Cache
}

type GithubAuthenticatorService interface {
	AuthenticateUser(ctx context.Context) (authenticaionmodels.GithubAccessToken, error)
}

const authenticaionCacheKey = "auth-key"

func NewGithubAuthenticator(githubauthenticationClient githubauthenticationclient.GithubAuthenticationClientService,
	cache *cache.Cache) GithubAuthenticator {
	return GithubAuthenticator{
		githubAuthenticationClient: githubauthenticationClient,
		cache:                      cache,
	}
}

func (g *GithubAuthenticator) AuthenticateUser(ctx context.Context) (authenticaionmodels.GithubAccessToken, error) {

	response, err := g.cache.GetOrCreate(authenticaionCacheKey, func(entry *cache.CacheEntry) (interface{}, error) {
		deviceCode, err := g.githubAuthenticationClient.GetDeviceCode(ctx)
		if err != nil {
			return authenticaionmodels.GithubAccessToken{}, err
		}

		displayUserInstructions(deviceCode)
		fmt.Print(deviceCode)
		result, err := g.githubAuthenticationClient.GetAccessToken(deviceCode, ctx)
		if err != nil {
			return authenticaionmodels.GithubAccessToken{}, err
		}

		entry.Expiration = time.Now().Add(time.Duration(deviceCode.ExpriesIn) * time.Second)

		return result, nil
	})
	if err != nil {
		return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error authenticating user: %w", err)
	}

	result, ok := response.(authenticaionmodels.GithubAccessToken)
	if !ok {
		return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error authenticating user, unable to read respone type: %w", err)
	}

	return result, nil
}

func displayUserInstructions(deviceResp authenticaionmodels.DeviceResposnse) {
	fmt.Printf("\n╭─────────────────────────────────────────╮\n")
	fmt.Printf("│          GitHub Authentication          │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n\n")
	fmt.Printf("1. Visit: %s\n", deviceResp.VerificationUrl)
	fmt.Printf("2. Enter code: %s\n", deviceResp.UserCode)
}
