package gitubauthenticationservice

import (
	"context"
	"fmt"

	githubauthenticationclient "github.com/RobsonDevCode/deepscan/internal/clients/githubAuthenticationClient"
	authenticaionmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/githubAuthentication"
)

type GithubAuthenticator struct {
	githubAuthenticationClient githubauthenticationclient.GithubAuthenticationClientService
}

type GithubAuthenticatorService interface {
	AuthenticateUser(ctx context.Context) (authenticaionmodels.GithubAccessToken, error)
}

func NewGithubAuthenticator(githubauthenticationClient githubauthenticationclient.GithubAuthenticationClientService) GithubAuthenticator {
	return GithubAuthenticator{
		githubAuthenticationClient: githubauthenticationClient,
	}
}

func (g *GithubAuthenticator) AuthenticateUser(ctx context.Context) (authenticaionmodels.GithubAccessToken, error) {
	deviceCode, err := g.githubAuthenticationClient.GetDeviceCode(ctx)
	if err != nil {
		return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error authenticating user: %w", err)
	}

	displayUserInstructions(deviceCode)
	fmt.Print(deviceCode)
	result, err := g.githubAuthenticationClient.GetAccessToken(deviceCode, ctx)
	if err != nil {
		return authenticaionmodels.GithubAccessToken{}, err
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
