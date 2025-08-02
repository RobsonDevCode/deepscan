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
	deviceCode, err := g.AuthenticateUser(ctx)
	if err != nil {
		return models.GithubAccessToken{}, fmt.Errorf("error authenticating user: %w", err)
	}

}
