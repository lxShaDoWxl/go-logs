package logs

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

func initZLog(conf *ConfigLogger) *zerolog.Logger {
	l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	if conf.JSON {
		l = zerolog.New(os.Stderr)
	}
	level, err := zerolog.ParseLevel(conf.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := l.Level(level).With().CallerWithSkipFrameCount(4).Timestamp().Logger()
	return &logger
}
