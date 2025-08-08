package githubauthenticationclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	authenticaionmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/githubAuthentication"
	"github.com/RobsonDevCode/deepscan/internal/configuration"
	"github.com/sony/gobreaker"
)

type GithubAuthenticationClient struct {
	client   *http.Client
	cb       *gobreaker.CircuitBreaker
	baseUrl  *url.URL
	cache    *cache.Cache
	clientId *string
}

type GithubAuthenticationClientService interface {
	GetDeviceCode(ctx context.Context) (authenticaionmodels.DeviceResposnse, error)
	GetAccessToken(deviceCode authenticaionmodels.DeviceResposnse, ctx context.Context) (authenticaionmodels.GithubAccessToken, error)
}

func NewGithubAuthenticationClient(config *configuration.Config, cache *cache.Cache) (*GithubAuthenticationClient, error) {
	client := &http.Client{
		Timeout: 1 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	cbSettings := gobreaker.Settings{
		Name:        "github-auth-client",
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

	baseUrl, err := url.Parse(config.GithubAuthenticationClientSettings.BaseUrl)
	if err != nil {
		return nil, fmt.Errorf("error parsing base url to a url type, %w", err)
	}

	cb := gobreaker.NewCircuitBreaker(cbSettings)
	return &GithubAuthenticationClient{
		client:   client,
		cb:       cb,
		baseUrl:  baseUrl,
		cache:    cache,
		clientId: &config.GithubAuthenticationClientSettings.ClientId,
	}, nil
}

func (c *GithubAuthenticationClient) GetDeviceCode(ctx context.Context) (authenticaionmodels.DeviceResposnse, error) {
	deviceCodeRequest := authenticaionmodels.DeviceCodeRequest{
		ClientId: *c.clientId,
		Scope:    "repo",
	}

	payload, err := json.Marshal(deviceCodeRequest)
	if err != nil {
		return authenticaionmodels.DeviceResposnse{}, fmt.Errorf("error marsheling device code request %w", err)
	}
	fmt.Printf("\npayload: '%s' \n", payload)

	cbResult, err := c.cb.Execute(func() (interface{}, error) {

		request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%slogin/device/code", c.baseUrl),
			bytes.NewBuffer(payload))
		fmt.Printf("URL: '%s'", request.URL)
		if err != nil {
			return authenticaionmodels.DeviceResposnse{}, fmt.Errorf("error creating http request for device code, %w", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")

		response, err := c.client.Do(request)
		if err != nil {
			return authenticaionmodels.DeviceResposnse{}, fmt.Errorf("error sending device code request, %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			authError, err := handleGithubAuthenticationError(response)
			if err != nil {
				return authenticaionmodels.DeviceResposnse{}, fmt.Errorf("error getting device code status %d:  %w", response.StatusCode, err)
			}

			return authenticaionmodels.DeviceCodeRequest{}, fmt.Errorf("error getting device code, client responded with %d, %s: %s", response.StatusCode, authError.Error, authError.ErrorDescription)
		}

		var result authenticaionmodels.DeviceResposnse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			return handleGithubAuthenticationError(response)
		}

		return result, nil
	})
	if err != nil {
		return authenticaionmodels.DeviceResposnse{}, err
	}

	result, ok := cbResult.(authenticaionmodels.DeviceResposnse)
	if !ok {
		return authenticaionmodels.DeviceResposnse{}, fmt.Errorf("unexpected response type when converting response")
	}

	return result, nil
}

func (c *GithubAuthenticationClient) GetAccessToken(deviceResp authenticaionmodels.DeviceResposnse, ctx context.Context) (authenticaionmodels.GithubAccessToken, error) {
	accessTokenRequest := authenticaionmodels.AccessTokenRequest{
		ClientId:   *c.clientId,
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		DeviceCode: deviceResp.DeviceCode,
	}
	fmt.Printf("\nDevice Code: %s", deviceResp.DeviceCode)

	payload, err := json.Marshal(accessTokenRequest)
	if err != nil {
		return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error marsheling device code request %w", err)
	}

	duration := time.Now().Add(time.Duration(deviceResp.ExpriesIn) * time.Second)

	response, err := c.executeAccessTokenRequest(payload, duration, ctx)
	if err != nil {
		return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error executing access token request: %w", err)
	}

	return response, nil
}

func (c *GithubAuthenticationClient) executeAccessTokenRequest(payload []byte, authTimeOut time.Time, ctx context.Context) (authenticaionmodels.GithubAccessToken, error) {
	status := "authorization_pending"
	fmt.Printf("auth timeout: %s", authTimeOut)

	for time.Now().Before(authTimeOut) && status != "completed" {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%slogin/oauth/access_token", c.baseUrl),
			bytes.NewBuffer(payload))
		if err != nil {
			return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error getting access token, failed to create http request: %w", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")

		response, err := c.client.Do(request)
		if err != nil {
			return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error sending access token request, %w", err)
		}
		defer response.Body.Close()

		if response.StatusCode != 200 {
			return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("access token request returned staus %d", response.StatusCode)
		}

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error couldnt get body bytes from authenticaion client: %w", err)
		}
		defer response.Body.Close()

		bodyString := string(bodyBytes)
		fmt.Printf("Json Body: %s", bodyString)

		var accessToken authenticaionmodels.GithubAccessToken
		var authError authenticaionmodels.AuthenticationError

		if err := json.Unmarshal(bodyBytes, &accessToken); err != nil {
			if authJsonError := json.Unmarshal(bodyBytes, &authError); authJsonError != nil {
				return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error unable to read token or auth error: %w", authJsonError)
			}
		}

		if accessToken.Token != "" {
			return accessToken, nil
		}

		if authError.Error == "" {
			if authJsonError := json.Unmarshal(bodyBytes, &authError); authJsonError != nil {
				return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error unable to read token or auth error: %w", authJsonError)
			}
		}

		fmt.Printf("\n Status: %s", authError.Error)
		if authError.Error == "slow_down" {
			time.Sleep(time.Duration(authError.Interval) * time.Second)
		}

		status = authError.Error
	}

	return authenticaionmodels.GithubAccessToken{}, fmt.Errorf("error token authentication period has expired please try again")
}

func handleGithubAuthenticationError(response *http.Response) (authenticaionmodels.AuthenticationError, error) {
	var clientError authenticaionmodels.AuthenticationError
	if err := json.NewDecoder(response.Body).Decode(&clientError); err != nil {
		return authenticaionmodels.AuthenticationError{}, fmt.Errorf("failed to read authetication error: %w", err)
	}

	return clientError, nil
}
