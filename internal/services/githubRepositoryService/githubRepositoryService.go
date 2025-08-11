package githubrepositoryservice

import (
	"context"
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/clients"
	"github.com/RobsonDevCode/deepscan/internal/clients/mapper"
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
	ghAccessToken, err := g.githubAuth.AuthenticateUser(ctx)
	if err != nil {
		return nil, err
	}

	ghRepos, err := g.githubClient.GetRepositories(ghAccessToken.Token, ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting repos: %w", err)
	}

	result := mapper.Map(ghRepos)
	for _, message := range result {
		fmt.Printf("\nRepo: %s SSH Url: %s", message.Name, message.SSHUrl)
	}

	return result, nil
}
