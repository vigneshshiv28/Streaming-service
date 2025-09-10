package logger

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type contextKey string

const LoggerKey contextKey = "logger"

func newLogger(logLevel string) *zerolog.Logger {

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		logger.Error().Msgf("Unknown level string %s defaulting to the info level", logLevel)
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)

	var logger zerolog.Logger

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"}

	output.FormatLevel = func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case "debug":
				l = colorize(ll, 36) // cyan
			case "info":
				l = colorize(ll, 34) // blue
			case "warn":
				l = colorize(ll, 33) // yellow
			case "error":
				l = colorize(ll, 31) // red
			case "fatal":
				l = colorize(ll, 35) // magenta
			case "panic":
				l = colorize(ll, 41) // white on red background
			default:
				l = colorize(ll, 37) // white
			}
		} else {
			if i == nil {
				l = colorize("???", 37) // white
			} else {
				lStr := strings.ToUpper(fmt.Sprintf("%s", i))
				if len(lStr) > 3 {
					lStr = lStr[:3]
				}
				l = lStr
			}
		}
		return fmt.Sprintf("| %s |", l)
	}

	logger = zerolog.New(output).With().Timestamp().Logger()

	return &logger

}

func InitLogger(logLevel string, ctx context.Context) (*zerolog.Logger, context.Context) {
	logger := newLogger(logLevel)
	ctx = context.WithValue(ctx, LoggerKey, logger)

	return logger, ctx
}

func colorize(s string, color int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, s)
}

func FromContex(ctx context.Context) *zerolog.Logger {
	log, ok := ctx.Value(LoggerKey).(*zerolog.Logger)

	if !ok {
		fallback := zerolog.New(os.Stderr).With().Timestamp().Logger()
		fallback.Error().Msg("Logger not found in context, using fallback")
		return &fallback
	}
	return log
}
