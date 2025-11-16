package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, arg ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Storage struct {
	db DB
}

var (
	ErrUserNotFound = errors.New("user not found")

	ErrTeamNotFound = errors.New("team not found")
	ErrTeamExists   = errors.New("team already exists")

	ErrPullRequestNotFound       = errors.New("pull request not found")
	ErrPullRequestExists         = errors.New("pull request already exists")
	ErrReassignMergedPullRequest = errors.New("cannot reassign on merge pull request")
	ErrUserNotAReviewer          = errors.New("user is not a reviewer of pull request")
	ErrNoCandidate               = errors.New("no active replacment candidadte in team")
)

func NewWithPool(p *pgxpool.Pool) *Storage {
	return &Storage{
		db: p,
	}
}

func NewWithTx(tx pgx.Tx) *Storage {
	return &Storage{
		db: tx,
	}
}

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

func (s *Storage) Close() {
	if pool, ok := s.db.(*pgxpool.Pool); ok {
		pool.Close()
	}
}

func Rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		slog.ErrorContext(
			ctx, "failed rollback transaction",
			"error", err)
	}
}
