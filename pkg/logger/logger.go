package logger

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var once sync.Once

// Init initializes the global logger. Should be called once at application startup.
func Init(isDevelopment bool) {
	once.Do(func() {
		if isDevelopment {
			log.Logger = log.Output(zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		} else {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		}
	})
}

// L returns the global logger instance.
func L() *zerolog.Logger {
	return &log.Logger
}

// Debug logs a debug message.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info logs an info message.
func Info() *zerolog.Event {
	return log.Info()
}

// Warn logs a warning message.
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs an error message.
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal logs a fatal message and exits.
func Fatal() *zerolog.Event {
	return log.Fatal()
}
