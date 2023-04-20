package util

import (
	"encoding/json"
	"errors"
	"fmt"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvProduction  = "production"
	EnvStage       = "stage"
	EnvIntegration = "integration"
	productionURL  = "https://api.openshift.com"
	stagingURL     = "https://api.stage.openshift.com"
	integrationURL = "https://api.integration.openshift.com"
)

var urlAliases = map[string]string{
	"production":   productionURL,
	"prod":         productionURL,
	"prd":          productionURL,
	productionURL:  productionURL,
	"staging":      stagingURL,
	"stage":        stagingURL,
	"stg":          stagingURL,
	stagingURL:     stagingURL,
	"integration":  integrationURL,
	"int":          integrationURL,
	integrationURL: integrationURL,
}

type Config struct {
	AccessToken  string   `json:"access_token,omitempty" doc:"Bearer access token."`
	ClientID     string   `json:"client_id,omitempty" doc:"OpenID client identifier."`
	ClientSecret string   `json:"client_secret,omitempty" doc:"OpenID client secret."`
	Insecure     bool     `json:"insecure,omitempty" doc:"Enables insecure communication with the server. This disables verification of TLS certificates and host names."`
	Password     string   `json:"password,omitempty" doc:"User password."`
	RefreshToken string   `json:"refresh_token,omitempty" doc:"Offline or refresh token."`
	Scopes       []string `json:"scopes,omitempty" doc:"OpenID scope. If this option is used it will replace completely the default scopes. Can be repeated multiple times to specify multiple scopes."`
	TokenURL     string   `json:"token_url,omitempty" doc:"OpenID token URL."`
	URL          string   `json:"url,omitempty" doc:"URL of the API gateway. The value can be the complete URL or an alias. The valid aliases are 'production', 'staging' and 'integration'."`
	User         string   `json:"user,omitempty" doc:"User name."`
	Pager        string   `json:"pager,omitempty" doc:"Pager command, for example 'less'. If empty no pager will be used."`
}

func CreateConnection() (*sdk.Connection, error) {
	ocmConfigError := "Unable to load OCM config\nLogin with 'ocm login' or set OCM_TOKEN, OCM_URL and OCM_REFRESH_TOKEN environment variables"

	connectionBuilder := sdk.NewConnectionBuilder()

	config, err := getOcmConfiguration(loadOCMConfig)
	if err != nil {
		return nil, errors.New(ocmConfigError)
	}

	connectionBuilder.Tokens(config.AccessToken, config.RefreshToken)
	if config.URL == "" {
		return nil, errors.New(ocmConfigError)
	}

	gatewayURL, ok := urlAliases[config.URL]
	if !ok {
		return nil, fmt.Errorf("Invalid OCM_URL found: %s\nValid URL aliases are: 'production', 'staging', 'integration'", config.URL)
	}
	connectionBuilder.URL(gatewayURL)

	connection, err := connectionBuilder.Build()

	if err != nil {
		if strings.Contains(err.Error(), "Not logged in, run the") {
			return nil, errors.New(ocmConfigError)
		}
		return nil, fmt.Errorf("failed to create OCM connection: %v", err)
	}

	return connection, nil
}

func getOcmConfiguration(ocmConfigLoader func() (*Config, error)) (*Config, error) {
	tokenEnv := os.Getenv("OCM_TOKEN")
	urlEnv := os.Getenv("OCM_URL")
	refreshTokenEnv := os.Getenv("OCM_REFRESH_TOKEN") // Unlikely to be set, but check anyway

	config := &Config{}

	// If missing required data, load from the config file.
	// We don't want to always load this, because the user might only use environment variables.
	if tokenEnv == "" || refreshTokenEnv == "" || urlEnv == "" {
		var fileConfigLoadError error
		config, fileConfigLoadError = ocmConfigLoader()
		if fileConfigLoadError != nil {
			return config, fmt.Errorf("could not load OCM configuration file")
		}
	}

	// Overwrite with set environment variables, to allow users to overwrite
	// their configuration file's variables
	if tokenEnv != "" {
		config.AccessToken = tokenEnv
	}
	if urlEnv != "" {
		config.URL = urlEnv
	}
	if refreshTokenEnv != "" {
		config.RefreshToken = refreshTokenEnv
	}

	return config, nil
}

func getOCMConfigLocation() (string, error) {
	if ocmconfig := os.Getenv("OCM_CONFIG"); ocmconfig != "" {
		return ocmconfig, nil
	}

	// Determine home directory to use for the legacy file path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".ocm.json")

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// Determine standard config directory
		configDir, err := os.UserConfigDir()
		if err != nil {
			return path, err
		}

		// Use standard config directory
		path = filepath.Join(configDir, "/ocm/ocm.json")
	}

	return path, nil
}

func loadOCMConfig() (*Config, error) {
	var err error

	file, err := getOCMConfigLocation()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		cfg := &Config{}
		err = nil
		return cfg, err
	}

	if err != nil {
		return nil, fmt.Errorf("can't check if config file '%s' exists: %v", file, err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("can't read config file '%s': %v", file, err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	cfg := &Config{}
	err = json.Unmarshal(data, cfg)

	if err != nil {
		return nil, fmt.Errorf("can't parse config file '%s': %v", file, err)
	}

	return cfg, nil
}

func GetCurrentEnv(connection *sdk.Connection) string {
	url := connection.URL()
	if strings.Contains(url, "stage") {
		return EnvStage
	}
	if strings.Contains(url, "integration") {
		return EnvIntegration
	}

	return EnvProduction
}
