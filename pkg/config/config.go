package config

import (
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/errors"

	"github.com/palantir/go-githubapp/githubapp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GithubHubAppConfig   githubapp.Config `yaml:"github_app_config" json:"github_app_config"`           //nolint:tagliatelle
	GHRepoTopic          string           `yaml:"github_repo_topic" json:"github_repo_topic"`           //nolint:tagliatelle
	GHRepoWhitelist      []string         `yaml:"github_repo_whitelist" json:"github_repo_whitelist"`   //nolint:tagliatelle
	SubFolderParallelism int              `yaml:"sub_folder_parallelism" json:"sub_folder_parallelism"` //nolint:tagliatelle
}

func LoadConfig() *Config {
	confLocation := os.Getenv("APP_CONF")
	if confLocation == "" {
		confLocation = "conf.yml"
	}

	data, err := ioutil.ReadFile(confLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config file")
	}
	var newConfig Config
	if err := yaml.Unmarshal(data, &newConfig); err != nil {
		log.Fatal().Err(err).Msg("Error Unmarshal config file")
	}

	errors := validateConfig(&newConfig)
	if len(errors) > 0 {
		log.Fatal().Errs("errors", errors).Msg("config validation error")
	}

	return &newConfig
}

func validateConfig(c *Config) []error {
	errs := []error{}

	if c.GithubHubAppConfig.WebURL == "" && c.GithubHubAppConfig.V3APIURL == "" && c.GithubHubAppConfig.V4APIURL == "" {
		errs = append(errs, errors.ConfigNotValidError("you must provide at least one of web_url / v3_api_url / v4_api_url in github_app_config field"))
	}

	if c.GithubHubAppConfig.App.IntegrationID == 0 || c.GithubHubAppConfig.App.WebhookSecret == "" || c.GithubHubAppConfig.App.PrivateKey == "" {
		errs = append(errs, errors.ConfigNotValidError("you must provide integration_id / webhook_secret / private_key in github_app_config.app field"))
	}

	if c.GithubHubAppConfig.OAuth.ClientID == "" || c.GithubHubAppConfig.OAuth.ClientSecret == "" {
		errs = append(errs, errors.ConfigNotValidError("you must provide client_id / client_secret in github_app_config.oauth field"))
	}

	if c.SubFolderParallelism == 0 {
		errs = append(errs, errors.ConfigNotValidError("you must provide sub_folder_parallelism field"))
	}
	return errs
}
