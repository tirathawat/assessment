package logs

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	debugMode bool
	logger    zerolog.Logger
}

func (l *Logger) createLogger() {
	var logger zerolog.Logger
	if l.debugMode {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp})
	} else {
		logger = zerolog.New(os.Stdout).With().Logger()
	}
	logger.Level(zerolog.InfoLevel)
	l.logger = logger
}

var logs *Logger

func init() {
	logs = &Logger{}
	logs.createLogger()
}

func Setup() {
	logs.createLogger()
}
