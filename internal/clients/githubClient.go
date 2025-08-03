package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	"github.com/RobsonDevCode/deepscan/internal/clients/models"
	"github.com/RobsonDevCode/deepscan/internal/configuration"
	"github.com/sony/gobreaker"
)

type GithubClientService interface {
	GetPackagesInfo(ecosystem string, packageAndVersions map[string]string, ctx context.Context) ([]models.ScannedPackage, error)
	GetDeviceCode(ctx context.Context) (models.DeviceResposnse, error)
	GetAccessToken(deviceCode models.DeviceResposnse, ctx context.Context) (models.GithubAccessToken, error)
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

func (c *GithubClient) GetDeviceCode(ctx context.Context) (models.DeviceResposnse, error) {
	deviceCodeRequest := models.DeviceCodeRequest{
		ClientId: *c.clientId,
		Scope:    "repo",
	}

	payload, err := json.Marshal(deviceCodeRequest)
	if err != nil {
		return models.DeviceResposnse{}, fmt.Errorf("error marsheling device code request %w", err)
	}

	cbResult, err := c.cb.Execute(func() (interface{}, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, "login/device/code",
			bytes.NewBuffer(payload))
		if err != nil {
			return models.DeviceResposnse{}, fmt.Errorf("error creating http request for device code, %w", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Acceps", "application/json")

		response, err := c.client.Do(request)
		if err != nil {
			return models.DeviceResposnse{}, fmt.Errorf("error sending device code request, %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			return nil, handleGithubClientError(response)
		}

		var result models.DeviceResposnse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return nil, handleGithubClientError(response.Request.Response)
		}

		return result, nil
	})
	if err != nil {
		return models.DeviceResposnse{}, err
	}

	result, ok := cbResult.(models.DeviceResposnse)
	if !ok {
		return models.DeviceResposnse{}, fmt.Errorf("unexpected response type when converting response")
	}

	return result, nil
}

func (c *GithubClient) GetAccessToken(deviceResp models.DeviceResposnse, ctx context.Context) (models.GithubAccessToken, error) {
	accessTokenRequest := models.AccessTokenRequest{
		ClientId:   *c.clientId,
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		DeviceCode: deviceResp.DeviceCode,
	}

	payload, err := json.Marshal(accessTokenRequest)
	if err != nil {
		return models.GithubAccessToken{}, fmt.Errorf("error marsheling device code request %w", err)
	}

	duration := time.Now().Add(time.Duration(deviceResp.ExpriesIn) * time.Second)

	response, err := c.executeAccessTokenRequest(payload, duration, ctx)
	if err != nil {
		return models.GithubAccessToken{}, fmt.Errorf("error executing access token request: %w", err)
	}

	if response.AccessToken == nil {
		result, err := c.handleAuthResponse(response, payload, duration, ctx)
		if err != nil {
			return models.GithubAccessToken{}, fmt.Errorf("error handling response: %w", err)
		}

		response.AccessToken = result.AccessToken
	}

	return *response.AccessToken, nil
}

func (c *GithubClient) executeAccessTokenRequest(payload []byte, authTimeout time.Time, ctx context.Context) (models.GithubAuthenticationResult, error) {

	fmt.Print("\nAuthenticating user....")
	if time.Now().After(authTimeout) {
		return models.GithubAuthenticationResult{
			AccessToken: nil,
			AuthError: &models.AuthenticationError{
				Error:            "expired_token",
				ErrorDescription: "auth window expired ",
			},
		}, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "login/oauth/access_token",
		bytes.NewBuffer(payload))
	if err != nil {
		return models.GithubAuthenticationResult{}, fmt.Errorf("error getting access token, failed to create http request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Acceps", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return models.GithubAuthenticationResult{}, fmt.Errorf("error sending access token request, %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		authError, err := handleGithubAuthenticationError(response)
		if err != nil {
			return models.GithubAuthenticationResult{}, err
		}

		return models.GithubAuthenticationResult{}, fmt.Errorf("error client responded with status code: %d message: %s", response.StatusCode, authError.ErrorDescription)
	}

	var accessToken *models.GithubAccessToken
	if err := json.NewDecoder(response.Body).Decode(&accessToken); err != nil {
		var authError *models.AuthenticationError

		if err := json.NewDecoder(response.Body).Decode(&authError); err != nil {
			return models.GithubAuthenticationResult{}, fmt.Errorf("error unable to read token or auth error")
		}

		return models.GithubAuthenticationResult{
			AccessToken: nil,
			AuthError:   authError,
		}, nil
	}

	//happy path
	result := models.GithubAuthenticationResult{
		AccessToken: accessToken,
		AuthError:   nil,
	}
	return result, nil
}

func (c *GithubClient) buildPackagesQuery(ecosystem string, packages map[string]string) string {
	baseUrl := fmt.Sprintf("advisories%s?ecosystem=%s", c.baseUrl, ecosystem)
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

func (c *GithubClient) handleAuthResponse(result models.GithubAuthenticationResult, payload []byte, duration time.Time, ctx context.Context) (models.GithubAuthenticationResult, error) {

	switch result.AuthError.Error {
	case "authorization_pending":
		fmt.Print("\nRetrying auth...")
		return c.executeAccessTokenRequest(payload, duration, ctx)
	case "slow_down":
		time.Sleep(1 * time.Second)
		return c.executeAccessTokenRequest(payload, duration, ctx)
	case "expired_token":
		return models.GithubAuthenticationResult{}, fmt.Errorf("device code expired")
	case "access_denied":
		return models.GithubAuthenticationResult{}, fmt.Errorf("user denied access")
	case "incorrect_client_credentials":
		return models.GithubAuthenticationResult{}, fmt.Errorf("invalid client_id")
	case "incorrect_device_code":
		return models.GithubAuthenticationResult{}, fmt.Errorf("invalid device_code")
	case "device_flow_disabled":
		return models.GithubAuthenticationResult{}, fmt.Errorf("device flow not enabled for this app")
	default:
		return models.GithubAuthenticationResult{}, fmt.Errorf("unknown error: %s", result.AuthError.Error)
	}
}

func handleGithubClientError(response *http.Response) error {
	var clientError models.Error
	if err := json.NewDecoder(response.Body).Decode(&clientError); err != nil {
		return fmt.Errorf("failed to read client error status code %d: %w", response.StatusCode, err)
	}

	return fmt.Errorf("client response error status: %d, %v", response.StatusCode, clientError)
}

func handleGithubAuthenticationError(response *http.Response) (models.AuthenticationError, error) {
	var clientError models.AuthenticationError
	if err := json.NewDecoder(response.Body).Decode(&clientError); err != nil {
		return models.AuthenticationError{}, fmt.Errorf("failed to read authetication error: %w", err)
	}

	return clientError, nil
}
