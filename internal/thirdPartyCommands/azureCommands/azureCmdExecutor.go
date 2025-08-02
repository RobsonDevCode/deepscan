package azurecommandExcecutor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	cmdmodels "github.com/RobsonDevCode/deepscan/internal/thirdPartyCommands/models"
)

type AzureCommandExcecutor struct {
	cache *cache.Cache
}

const key = "repos"
const uiProject = "ui"
const pocProject = "poc"

type AzureCommandService interface {
	GetRepos(orgUrl string, project string) ([]cmdmodels.Repository, error)
}

func NewAzureCommandExcecutor(cache *cache.Cache) *AzureCommandExcecutor {
	return &AzureCommandExcecutor{
		cache: cache,
	}
}

func (a *AzureCommandExcecutor) GetRepos(orgUrl string, project string) ([]cmdmodels.Repository, error) {
	response, err := a.cache.GetOrCreate(key, time.Minute*10, func() (interface{}, error) {
		repoCommand := exec.Command("az", "repos", "list", "--org", orgUrl, "-p", project)
		repoCommand.Env = os.Environ()

		output, err := repoCommand.CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "you need to run the login command") {
				loginCmd := exec.Command("az", "login")
				_, err := loginCmd.Output()
				if err != nil {
					return nil, fmt.Errorf("error logging into to azure, %v", err)
				}

				repoCommand := exec.Command("az", "repos", "list", "--org", orgUrl, "-p", project)
				repoCommand.Env = os.Environ()

				secondTry, err := repoCommand.CombinedOutput()
				if err != nil {
					return nil, err
				}

				output = secondTry

			} else {
				return nil, fmt.Errorf("error getting repos: %v\n Output: %s", err, string(output))
			}
		}

		var repos []cmdmodels.Repository
		if err := json.Unmarshal(output, &repos); err != nil {
			return nil, fmt.Errorf("error reading json result: %v", err)
		}

		var result []cmdmodels.Repository

		for _, repo := range repos {
			if repo.IsDisabled {
				continue
			}
			result = append(result, repo)
		}

		return result, nil
	})

	if err != nil {
		return nil, err
	}

	result, ok := response.([]cmdmodels.Repository)
	if !ok {
		return nil, fmt.Errorf("unexpected response type, could not map correctly")
	}

	return result, err
}
