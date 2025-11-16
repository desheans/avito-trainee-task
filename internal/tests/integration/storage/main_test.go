package storage

import (
	"context"
	"log"
	"os"
	"testing"

	"avito-trainee-task/internal/storage/postgres"
	"avito-trainee-task/internal/tests"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

const (
	schemeMigrationsPath = "../../../../migrations/"
)

var (
	pool *pgxpool.Pool
	ctx  = context.Background()
)

func TestMain(m *testing.M) {
	var err error
	pool, err = tests.CreateTestPool(ctx, schemeMigrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func setupTestStorage(t *testing.T) (pgx.Tx, *postgres.Storage, func()) {
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	storage := postgres.NewWithTx(tx)

	cleanup := func() {
		_ = tx.Rollback(ctx)
	}

	return tx, storage, cleanup
}
