package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	v1 "avito-trainee-task/internal/controller/http/v1"
	"avito-trainee-task/internal/migration"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	pg "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func CreateTestPool(ctx context.Context, migrationPath string) (*pgxpool.Pool, error) {
	ctr, err := pg.Run(
		ctx,
		"postgres:16-alpine",
		pg.WithDatabase("postgres"),
		pg.WithUsername("postgres"),
		pg.WithPassword("postgres_password"),
		pg.BasicWaitStrategies(),
		pg.WithSQLDriver("pgx"),
	)
	if err != nil {
		return nil, fmt.Errorf("CreateTestPool failed to run postgres container: %v", err)
	}

	dbURL, err := ctr.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateTestPool failed to obtain connection string: %v", err)
	}

	if err = migration.Migrate(migrationPath, dbURL); err != nil {
		return nil, fmt.Errorf("CreateTestPool failed to migrate scheme: %v", err)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("CreateTestPool failed to create pool: %v", err)
	}

	return pool, nil
}

func CreatePostgresStorage(ctx context.Context, migrationPath string) (*postgres.Storage, error) {
	pool, err := CreateTestPool(ctx, migrationPath)
	if err != nil {
		return nil, err
	}

	return postgres.NewWithPool(pool), nil
}

func StartServer(s v1.Storage) (string, func(), error) {
	e := echo.New()
	h := v1.NewHandler(s)
	h.RegisterRoutes(e)

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, fmt.Errorf("failed to listen: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	serverURL := fmt.Sprintf("http://localhost:%d", port)

	go func() {
		if err := e.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)

	resp, err := http.Get(serverURL + "/")
	if err != nil {
		return "", nil, fmt.Errorf("server not responding: %v", err)
	}
	resp.Body.Close()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = e.Shutdown(ctx)
	}

	return serverURL, cleanup, nil
}
