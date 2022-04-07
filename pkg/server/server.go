package server

import (
	"fmt"
	"net/http"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/config"
	"github.com/terraform-tools/terraform-checker/pkg/github"
)

const ListeningPort = 8000

func StartServer() {
	config := config.LoadConfig()

	cc, err := githubapp.NewDefaultCachingClientCreator(config.GithubHubAppConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating githubapp client")
	}

	mainHandler := &github.CheckHandler{Client: cc, Config: config}
	webhookHandler := githubapp.NewDefaultEventDispatcher(config.GithubHubAppConfig, mainHandler)

	mux := http.NewServeMux()
	mux.Handle(githubapp.DefaultWebhookRoute, webhookHandler)
	mux.HandleFunc("/ping", PingHandler)
	log.Info().Msgf("Starting webserver, listening :%d", ListeningPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", ListeningPort), mux)
	if err != nil {
		log.Error().Err(err).Msg("Error creating webserver")
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
