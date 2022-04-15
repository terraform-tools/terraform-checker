package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger() {
	debug := os.Getenv("TF_CHECKER_DEBUG")
	log.Logger = log.With().Caller().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug != "" && (debug == "1" || debug == "true") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
