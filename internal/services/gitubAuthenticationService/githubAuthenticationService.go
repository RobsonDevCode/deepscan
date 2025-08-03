package gitubauthenticationservice

import (
	"context"
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/clients"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
)

type GithubAuthenticator struct {
	githubClient clients.GithubClientService
}

type GithubAuthenticatorService interface {
	AuthenticateUser(ctx context.Context) (models.GithubAccessToken, error)
}

func NewGithubAuthenticator(githubClient clients.GithubClientService) GithubAuthenticator {
	return GithubAuthenticator{
		githubClient: githubClient,
	}
}

func (g *GithubAuthenticator) AuthenticateUser(ctx context.Context) (models.GithubAccessToken, error) {
	deviceCode, err := g.githubClient.GetDeviceCode(ctx)
	if err != nil {
		return models.GithubAccessToken{}, fmt.Errorf("error authenticating user: %w", err)
	}

	displayUserInstructions(deviceCode)

	result, err := g.githubClient.GetAccessToken(deviceCode, ctx)
	if err != nil {
		return models.GithubAccessToken{}, err
	}

	return result, nil
}

func displayUserInstructions(deviceResp models.DeviceResposnse) {
	fmt.Printf("\n╭─────────────────────────────────────────╮\n")
	fmt.Printf("│          GitHub Authentication          │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n\n")
	fmt.Printf("1. Visit: %s\n", deviceResp.VerificationUrl)
	fmt.Printf("2. Enter code: %s\n", deviceResp.UserCode)
}
