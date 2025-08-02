package githubrepositoryservice

import (
	"context"

	"github.com/RobsonDevCode/deepscan/internal/clients"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

type GitRepositoryService interface {
	GetRepos(profile string, ctx context.Context) ([]cmdmodels.Repository, error)
}

type GitHubRepositoryRetrival struct {
	githubClient clients.GithubClientService
}

func NewGithubRepositoryRetrivalService(githubClient clients.GithubClientService) GitHubRepositoryRetrival {
	return GitHubRepositoryRetrival{
		githubClient: githubClient,
	}
}

func (g *GitHubRepositoryRetrival) GetRepos(profile string, ctx context.Context) ([]cmdmodels.Repository, error) {

	return nil, nil
}
