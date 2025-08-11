package gitubauthenticationservice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	cache "github.com/RobsonDevCode/deepscan/internal/caching"
	githubauthenticationclient "github.com/RobsonDevCode/deepscan/internal/clients/githubAuthenticationClient"
	authenticationmodels "github.com/RobsonDevCode/deepscan/internal/clients/models/githubAuthentication"
	setupservice "github.com/RobsonDevCode/deepscan/internal/services/setupService"
)

type GithubAuthenticator struct {
	githubAuthenticationClient githubauthenticationclient.GithubAuthenticationClientService
	cache                      *cache.Cache
}

type GithubAuthenticatorService interface {
	AuthenticateUser(ctx context.Context) (authenticationmodels.GithubAccessToken, error)
}

const (
	authenticaionCacheKey = "auth-key"
	filePath              = "configuration/user_setting.json"
	tmpFilePath           = "configuration/user_setting_tmp.json"
)

func NewGithubAuthenticator(githubauthenticationClient githubauthenticationclient.GithubAuthenticationClientService,
	cache *cache.Cache) GithubAuthenticator {
	return GithubAuthenticator{
		githubAuthenticationClient: githubauthenticationClient,
		cache:                      cache,
	}
}

func (g *GithubAuthenticator) AuthenticateUser(ctx context.Context) (authenticationmodels.GithubAccessToken, error) {

	accessToken, err := getLocalAccessToken()
	if err != nil {
		return authenticationmodels.GithubAccessToken{}, err
	}
	if accessToken != nil {
		return *accessToken, nil
	}

	response, err := g.cache.GetOrCreate(authenticaionCacheKey, func(entry *cache.CacheEntry) (interface{}, error) {
		deviceCode, err := g.githubAuthenticationClient.GetDeviceCode(ctx)
		if err != nil {
			return authenticationmodels.GithubAccessToken{}, err
		}

		displayUserInstructions(deviceCode)
		result, err := g.githubAuthenticationClient.GetAccessToken(deviceCode, ctx)
		if err != nil {
			return authenticationmodels.GithubAccessToken{}, err
		}

		entry.Expiration = time.Now().Add(time.Duration(deviceCode.ExpriesIn) * time.Second)
		return result, nil
	})
	if err != nil {
		return authenticationmodels.GithubAccessToken{}, fmt.Errorf("error authenticating user: %w", err)
	}

	deviceCode, ok := response.(authenticationmodels.DeviceResposnse)
	if !ok {
		return authenticationmodels.GithubAccessToken{}, fmt.Errorf("error authenticating user, unable to read respone type: %w", err)
	}

	*accessToken, err = g.githubAuthenticationClient.GetAccessToken(deviceCode, ctx)

	if err != nil {
		return authenticationmodels.GithubAccessToken{}, fmt.Errorf("error getting access token: %w", err)
	}

	if err := setLocalAccessToken(*accessToken); err != nil {
		fmt.Print("error saving access token locally: %w", err) //log but dont fail
		return *accessToken, nil
	}

	return *accessToken, nil
}

func displayUserInstructions(deviceResp authenticationmodels.DeviceResposnse) {
	fmt.Printf("\n╭─────────────────────────────────────────╮\n")
	fmt.Printf("│          GitHub Authentication          │\n")
	fmt.Printf("╰─────────────────────────────────────────╯\n\n")
	fmt.Printf("1. Visit: %s\n", deviceResp.VerificationUrl)
	fmt.Printf("2. Enter code: %s\n", deviceResp.UserCode)
}

func getLocalAccessToken() (*authenticationmodels.GithubAccessToken, error) {
	userSettings, err := setupservice.GetUserSettings()
	if err != nil {
		return nil, fmt.Errorf("error getting user setting: %w", err)
	}

	if userSettings.AccessToken != nil {
		return userSettings.AccessToken, nil
	}

	return nil, nil //no error but no access token so user hasnt authenticated before
}

func setLocalAccessToken(accessToken authenticationmodels.GithubAccessToken) error {
	userSettings, err := setupservice.GetUserSettings()
	if err != nil {
		return fmt.Errorf("error getting user settings while creating local access token: %w", err)
	}

	userSettings.AccessToken = &accessToken
	updatedJson, _ := json.Marshal(userSettings)

	if err := os.WriteFile(tmpFilePath, updatedJson, 0644); err != nil {
		return err
	}

	os.Remove(filePath)
	os.Rename(tmpFilePath, filePath)

	return nil
}
