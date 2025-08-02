package repositoryreaderservice

import (
	"fmt"

	"github.com/RobsonDevCode/deepscan/internal/configuration"
	supportedproviders "github.com/RobsonDevCode/deepscan/internal/constants/supportedProviders"
	azurecommandExcecutor "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/azureCommands"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

type RepositoryReaderFacade interface {
	GetRepos(userSettings configuration.UsersSettings) ([]cmdmodels.Repository, error)
}

type RepositoryReaderService struct {
	azureCmds azurecommandExcecutor.AzureCommandService
}

func NewRepositoryReaderService(azureCmds azurecommandExcecutor.AzureCommandService) RepositoryReaderService {
	return RepositoryReaderService{
		azureCmds: azureCmds,
		//TODO add github support
	}
}

func (r *RepositoryReaderService) GetRepos(userSettings configuration.UsersSettings) ([]cmdmodels.Repository, error) {
	switch userSettings.Provider {
	case supportedproviders.Azure:
		return r.azureCmds.GetRepos(userSettings.OrganizationUrl, userSettings.Profile)

	case supportedproviders.Github:
		//TODO
		return nil, fmt.Errorf("not yet implimented")

	default:
		return nil, fmt.Errorf("non supported provider provided")
	}
}
