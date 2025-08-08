package githubrepositoryservice

import (
	"context"
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/clients"
	gitubauthenticationservice "github.com/RobsonDevCode/deepscan/internal/services/gitubAuthenticationService"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

type GitRepositoryService interface {
	GetRepos(profile string, ctx context.Context) ([]cmdmodels.Repository, error)
}

type GitHubRepositoryRetrival struct {
	githubClient clients.GithubClientService
	githubAuth   gitubauthenticationservice.GithubAuthenticatorService
}

func NewGithubRepositoryRetrivalService(githubClient clients.GithubClientService, githubAuth gitubauthenticationservice.GithubAuthenticatorService) GitHubRepositoryRetrival {
	return GitHubRepositoryRetrival{
		githubClient: githubClient,
		githubAuth:   githubAuth,
	}
}

func (g *GitHubRepositoryRetrival) GetRepos(profile string, ctx context.Context) ([]cmdmodels.Repository, error) {
	result, err := g.githubAuth.AuthenticateUser(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Token: %s", result.Token)

	return nil, nil
}
