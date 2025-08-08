package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	githubreposmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/repos"
	"github.com/RobsonDevCode/deepscan/internal/configuration"
	"github.com/sony/gobreaker"
)

type GithubClientService interface {
	GetPackagesInfo(ecosystem string, packageAndVersions map[string]string, ctx context.Context) ([]models.ScannedPackage, error)
}

type GithubClient struct {
	client              *http.Client
	cb                  *gobreaker.CircuitBreaker
	baseUrl             *url.URL
	cache               *cache.Cache
	personalAccessToken *string
	clientId            *string
}

func NewGithubClient(config *configuration.Config, cache *cache.Cache) (*GithubClient, error) {
	client := &http.Client{
		Timeout: 1 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	cbSettings := gobreaker.Settings{
		Name:        "git-client",
		MaxRequests: 5,
		Interval:    3 * time.Second,
		Timeout:     20 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit breaker state changed from %v to %v\n", from, to)
		},
	}

	baseUrl, err := url.Parse(config.GithubClientSettings.BaseUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing base url to a url type, %w", err)
	}
	cb := gobreaker.NewCircuitBreaker(cbSettings)

	return &GithubClient{
		client:              client,
		cb:                  cb,
		baseUrl:             baseUrl,
		cache:               cache,
		personalAccessToken: &config.GithubClientSettings.PAT,
		clientId:            &config.GithubClientSettings.ClientId,
	}, nil
}

func (c *GithubClient) GetPackagesInfo(ecosystem string, packageAndVersions map[string]string, ctx context.Context) ([]models.ScannedPackage, error) {
	if len(packageAndVersions) == 0 || packageAndVersions == nil {
		fmt.Print("No packages on project")
		return nil, nil
	}

	url := c.buildPackagesQuery(ecosystem, packageAndVersions)
	if url == "" {
		return nil, nil
	}

	cbResult, err := c.cb.Execute(func() (interface{}, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %w", err)
		}

		request.Header.Set("Authorization", "token "+*c.personalAccessToken)

		response, err := c.client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("client response error: %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			return nil, handleGithubClientError(response)
		}

		var results []models.ScannedPackage
		if err := json.NewDecoder(response.Body).Decode(&results); err != nil {
			return nil, handleGithubClientError(request.Response)
		}

		return results, nil
	})

	if err != nil {
		return nil, err
	}

	results, ok := cbResult.([]models.ScannedPackage)
	if !ok {
		return nil, fmt.Errorf("unexpected response type when converting response")
	}

	for i := range results {
		for j := range results[i].Vulnerabilities {
			packageVersion := packageAndVersions[results[i].Vulnerabilities[j].Package.Name]
			results[i].Vulnerabilities[j].CurrentVersion = packageVersion
		}
	}

	return results, nil
}

func (c *GithubClient) GetRepositories(accessToken string, ctx context.Context) (githubreposmodels.GithubRepositoryResult, error) {
	url := fmt.Sprintf("%suser/repos", c.baseUrl)

	cbResult, cbErr := c.cb.Execute(func() (interface{}, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating http request %s", err)
		}

		request.Header.Set("Authorization", "Bearer "+url)

		response, err := c.client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("client respoded with status %d", response.StatusCode)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			var errorMessage githubreposmodels.ErrorResponse
			if err := json.NewDecoder(response.Body).Decode(&errorMessage); err != nil {
				return nil, fmt.Errorf("error client responded with %d and failed to read error message", response.StatusCode)
			}
			return nil, fmt.Errorf("error github client responded with %d message: %s", response.StatusCode)
		}

		//TODO finish this
		return nil, nil
	})
	if cbErr != nil {
		return githubreposmodels.GithubRepositoryResult{}, cbErr
	}

	result, ok := cbResult.(githubreposmodels.GithubRepositoryResult)
	if !ok {
		return githubreposmodels.GithubRepositoryResult{}, fmt.Errorf("unexpected response type when converting response")
	}

	return result, nil
}

func (c *GithubClient) buildPackagesQuery(ecosystem string, packages map[string]string) string {
	baseUrl := fmt.Sprintf("%sadvisories?ecosystem=%s", c.baseUrl, ecosystem)
	var urlBuilder strings.Builder
	urlBuilder.WriteString(baseUrl)

	firstPackage := true
	for packageName, version := range packages {
		if firstPackage {
			urlBuilder.WriteString("&affects=")
			firstPackage = false
		} else {
			urlBuilder.WriteString(",")
		}

		urlBuilder.WriteString(url.QueryEscape(packageName))
		if version != "" && version != "0.0.0" {
			urlBuilder.WriteString("@")
			urlBuilder.WriteString(url.QueryEscape(version))
		}
	}

	completeString := urlBuilder.String()
	if completeString == baseUrl {
		fmt.Print("\n No packages to scan on project")
		return ""
	}

	return urlBuilder.String()
}

func handleGithubClientError(response *http.Response) error {
	var clientError models.Error
	if err := json.NewDecoder(response.Body).Decode(&clientError); err != nil {
		return fmt.Errorf("failed to read client error status code %d: %w", response.StatusCode, err)
	}

	return fmt.Errorf("client response error status: %d, %v", response.StatusCode, clientError)
}
