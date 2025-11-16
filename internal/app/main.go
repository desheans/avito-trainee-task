package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"avito-trainee-task/config"
	v1 "avito-trainee-task/internal/controller/http/v1"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run(ctx context.Context, cfg *config.Config) error {
	logger := setupLogger(cfg)
	slog.SetDefault(logger)

	pool, err := postgres.NewPool(ctx, cfg.Postgres.PGURL)
	if err != nil {
		panic(err)
	}
	s := postgres.NewWithPool(pool)
	defer s.Close()

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${method} ${uri} ${status} ${error}\n",
	}))
	h := v1.NewHandler(s)
	h.RegisterRoutes(e)

	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return e.Shutdown(ctx)
}
