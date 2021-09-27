package server

import (
	"net/http"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/rs/zerolog/log"
	"github.com/terraform-tools/terraform-checker/pkg/github"
)

func StartServer() {
	config := loadConfig()

	cc, err := githubapp.NewDefaultCachingClientCreator(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating githubapp client")
	}
	prCommentHandler := &github.CheckHandler{Client: cc}

	webhookHandler := githubapp.NewDefaultEventDispatcher(config, prCommentHandler)

	mux := http.NewServeMux()
	mux.Handle(githubapp.DefaultWebhookRoute, webhookHandler)
	log.Print("Starting webserver")
	err = http.ListenAndServe(":8000", mux)
	if err != nil {
		log.Error().Err(err).Msg("Error creating webserver")
	}
}
