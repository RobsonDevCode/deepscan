package configuration

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const FilePath = "configuration/configuration.yaml"

type Config struct {
	GithubClientSettings GethubClinetSettings `yaml:"github_client_settings"`
}

type GethubClinetSettings struct {
	BaseUrl string `yaml:"base_url"`
	PAT     string `yaml:"personal_access_token"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile(FilePath)
	if err != nil {
		return nil, fmt.Errorf("configuration error: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling configuration: %w", err)
	}

	return &config, nil
}
