package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/config"
	"github.com/terraform-tools/terraform-checker/pkg/github"
)

const (
	ListeningPort            = 8000
	ReadHeaderTimeoutSeconds = 3
)

func StartServer() {
	config := config.LoadConfig()

	cc, err := githubapp.NewDefaultCachingClientCreator(config.GithubHubAppConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating githubapp client")
	}

	mainHandler := &github.CheckHandler{Client: cc, Config: config}
	mainHandler.Init()

	webhookHandler := githubapp.NewDefaultEventDispatcher(config.GithubHubAppConfig, mainHandler)

	mux := http.NewServeMux()
	mux.Handle(githubapp.DefaultWebhookRoute, webhookHandler)
	mux.HandleFunc("/ping", PingHandler)
	log.Info().Msgf("Starting webserver, listening :%d", ListeningPort)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", ListeningPort),
		ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
		Handler:           mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Error().Err(err).Msg("Error creating webserver")
	}
	if err != nil {
		panic(err)
	}
}

func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
