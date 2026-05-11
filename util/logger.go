package util

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/shivangp0208/bank_application/config"
)

var logger zerolog.Logger
var utilConfig config.Config

func init() {
	utilConfig = config.GetConfig()
}

func GetLogger() *zerolog.Logger {
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return &logger
}
