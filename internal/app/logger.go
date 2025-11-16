package app

import (
	"log/slog"
	"os"

	"avito-trainee-task/config"
)

func setupLogger(cfg *config.Config) *slog.Logger {
	if cfg.Env == config.Development {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
