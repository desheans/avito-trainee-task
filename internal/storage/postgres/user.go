package postgres

import (
	"context"
	"errors"
	"fmt"

	"avito-trainee-task/internal/api"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetReview(ctx context.Context, userId string) ([]*api.PullRequestShort, error) {
	const op = "postgres.GetReview"
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v failed to begin transaction: %w", op, err)
	}
	defer Rollback(ctx, tx)

	exists, err := s.IsUserExists(ctx, tx, userId)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, ErrUserNotFound
	}

	sql := `SELECT author_id, pull_request_id, pull_request_name, status FROM pull_requests
	WHERE $1 = ANY(assigned_reviewers)`
	rows, err := tx.Query(ctx, sql, userId)
	if err != nil {
		return nil, fmt.Errorf("%v failed to query pull requests: %w", op, err)
	}

	prs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*api.PullRequestShort, error) {
		var pr api.PullRequestShort
		err := row.Scan(&pr.AuthorId, &pr.PullRequestId, &pr.PullRequestName, &pr.Status)
		if err != nil {
			return nil, fmt.Errorf("%v failed to scan row: %w", op, err)
		}
		return &pr, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to collect rows: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%v failed to commit transaction: %w", op, err)
	}

	return prs, nil
}

func (s *Storage) SetIsActive(ctx context.Context, UserId string, isActive bool) (*api.User, error) {
	const op = "postgres.SetIsActive"
	sql := "UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING *"

	var user api.User
	err := s.db.QueryRow(ctx, sql, isActive, UserId).Scan(
		&user.UserId,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%v error to query user: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetUsersStats(ctx context.Context) (*api.AssignmentCountStat, error) {
	const op = "postgres.GetStats"
	sql := `
	SELECT 
		user_id, 
		COUNT(*) as assignment_count 
	FROM pull_requests, 
    	unnest(assigned_reviewers) as user_id
	GROUP BY user_id
	ORDER BY assignment_count DESC`

	rows, err := s.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("%v failed to query: %w", op, err)
	}

	stat, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (api.AssignmentCount, error) {
		var ac api.AssignmentCount
		return ac, row.Scan(&ac.UserId, &ac.AssignmentCount)
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to collect rows: %w", op, err)
	}

	return &api.AssignmentCountStat{
		Stats: stat,
	}, nil
}

func (s *Storage) GetTeamNameByUserId(ctx context.Context, tx pgx.Tx, userId string) (string, error) {
	sql := "SELECT team_name FROM users WHERE user_id = $1"
	var authorTeam string
	if err := tx.QueryRow(ctx, sql, userId).Scan(&authorTeam); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("postgres.GetTeamNameByUserId failed to query row: %w", err)
	} else if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	return authorTeam, nil
}

func (s *Storage) IsUserExists(ctx context.Context, tx pgx.Tx, userId string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`
	var ok bool
	if err := tx.QueryRow(ctx, sql, userId).Scan(&ok); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ok, fmt.Errorf("postgres.IsUserExists failed to query row: %w", err)
	}
	return ok, nil
}
