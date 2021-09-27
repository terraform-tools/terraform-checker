package logger

import "github.com/rs/zerolog/log"

func SetupLogger() {
	log.Logger = log.With().Caller().Logger()
}
