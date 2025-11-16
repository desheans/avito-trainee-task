package migration

import (
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(sourcePath, databaseUrl string) error {
	databaseUrl = strings.Replace(databaseUrl, "postgres://", "pgx5://", 1)

	m, err := migrate.New(
		"file://"+sourcePath,
		databaseUrl,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
