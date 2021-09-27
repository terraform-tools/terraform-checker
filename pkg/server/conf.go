package server

import (
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/palantir/go-githubapp/githubapp"
	"gopkg.in/yaml.v2"
)

func loadConfig() githubapp.Config {
	confLocation := os.Getenv("APP_CONF")
	if confLocation == "" {
		confLocation = "conf.yml"
	}

	data, err := ioutil.ReadFile(confLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config file")
	}
	var config githubapp.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Msg("Error Unmarshal config file")
	}

	return config
}
