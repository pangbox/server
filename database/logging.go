package database

import (
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

func SetLogger(log zerolog.Logger) {
	goose.SetLogger(zerologGooseLogger{log})
}

type zerologGooseLogger struct {
	log zerolog.Logger
}

func (z zerologGooseLogger) Fatalf(format string, v ...interface{}) {
	z.log.Fatal().Msgf(strings.TrimSuffix(format, "\n"), v...)
}

func (z zerologGooseLogger) Printf(format string, v ...interface{}) {
	z.log.Info().Msgf(strings.TrimSuffix(format, "\n"), v...)
}
