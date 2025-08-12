package repositoryreaderservice

import (
	"context"
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/configuration"
	supportedproviders "github.com/RobsonDevCode/deepscan/internal/constants/supportedProviders"
	githubrepositoryservice "github.com/RobsonDevCode/deepscan/internal/services/githubRepositoryService"
	azurecommandExcecutor "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/azureCommands"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

type RepositoryReaderFacade interface {
	GetRepos(userSettings configuration.UsersSettings, ctx context.Context) ([]cmdmodels.Repository, error)
}

type RepositoryReaderService struct {
	azureCmds         azurecommandExcecutor.AzureCommandService
	githubRepoService githubrepositoryservice.GitRepositoryService
}

func NewRepositoryReaderService(azureCmds azurecommandExcecutor.AzureCommandService,
	githubRepoService githubrepositoryservice.GitRepositoryService) RepositoryReaderService {
	return RepositoryReaderService{
		azureCmds:         azureCmds,
		githubRepoService: githubRepoService,
	}
}

func (r *RepositoryReaderService) GetRepos(userSettings configuration.UsersSettings, ctx context.Context) ([]cmdmodels.Repository, error) {
	switch userSettings.Provider {
	case supportedproviders.Azure:
		return r.azureCmds.GetRepos(fmt.Sprintf("%v", userSettings.OrganizationUrl), userSettings.Profile)

	case supportedproviders.Github:
		return r.githubRepoService.GetRepos(userSettings.Profile, ctx)

	default:
		return nil, fmt.Errorf("non supported provider provided")
	}
}
