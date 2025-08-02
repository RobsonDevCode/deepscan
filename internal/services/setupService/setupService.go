package setupservice

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RobsonDevCode/deepscan/internal/configuration"
	supportedproviders "github.com/RobsonDevCode/deepscan/internal/constants/supportedProviders"
)

const filePath = "configuration/user_setting.json"

func CreateSetupFile(url string, profile string) error {

	if _, err := os.Stat(filePath); os.IsExist(err) {
		return fmt.Errorf("user settings already exist")
	}

	var provider string
	if strings.Contains(url, supportedproviders.GithubUrl) {
		provider = supportedproviders.Github
	}
	if strings.Contains(url, supportedproviders.AzureUrl) {
		provider = supportedproviders.Azure
	}

	userSettings := configuration.UsersSettings{
		OrganizationUrl: url,
		Profile:         profile,
		Provider:        provider,
	}

	jsonData, err := json.Marshal(userSettings)
	if err != nil {
		return fmt.Errorf("error marsheling json, %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing file at %s, %w", filePath, err)
	}

	return nil
}

func GetUserSettings() (*configuration.UsersSettings, error) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read user settings: %w", err)
	}

	var userSettings configuration.UsersSettings
	err = json.Unmarshal(jsonData, &userSettings)
	if err != nil {
		return nil, fmt.Errorf("error unmarsheling user settings %w", err)
	}

	if userSettings.OrganizationUrl == "" || userSettings.Profile == "" {
		return nil, fmt.Errorf("error, validating set up please insure the set up command has been ran")
	}

	return &userSettings, nil
}
