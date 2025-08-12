package mapper

import (
	githubreposmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/repos"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

func Map(githubRepos []githubreposmodels.GithubRepository) []cmdmodels.Repository {
	result := make([]cmdmodels.Repository, 0)

	for _, githubRepo := range githubRepos {
		repo := cmdmodels.Repository{
			SSHUrl:     githubRepo.CloneUrl,
			Name:       githubRepo.Name,
			IsDisabled: false,
		}
		result = append(result, repo)
	}

	return result
}
