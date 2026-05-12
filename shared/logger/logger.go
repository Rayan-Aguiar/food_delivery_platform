package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	ServiceName string
	Level       string
}

func New(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler).With(
		"service", cfg.ServiceName,
	)
}
