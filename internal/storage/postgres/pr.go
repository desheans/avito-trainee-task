package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"avito-trainee-task/internal/api"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) Merge(ctx context.Context, pullRequestId string) (*api.PullRequest, error) {
	sql := `UPDATE pull_requests 
		SET status = $1 
	WHERE pull_request_id = $2 
	RETURNING pull_request_id, pull_request_name, status, author_id, assigned_reviewers, createdAt, mergedAt`

	var request api.PullRequest
	err := s.db.QueryRow(ctx, sql, api.PullRequestShortStatusMERGED, pullRequestId).Scan(
		&request.PullRequestId,
		&request.PullRequestName,
		&request.Status,
		&request.AuthorId,
		&request.AssignedReviewers,
		&request.CreatedAt,
		&request.MergedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPullRequestNotFound
	} else if err != nil {
		return nil, fmt.Errorf("postgres.Merge failed query row: %w", err)
	}

	return &request, nil
}

func (s *Storage) Reassign(ctx context.Context, pullRequestId, userId string) (*api.PullRequest, string, error) {
	const op = "postgres.Reassign"
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%v failed to begin transaction: %w", op, err)
	}
	defer Rollback(ctx, tx)

	pr, err := s.GetPullRequest(ctx, tx, pullRequestId)
	if err != nil {
		return nil, "", err
	} else if pr.Status == api.PullRequestStatusMERGED {
		return nil, "", ErrReassignMergedPullRequest
	} else if !slices.Contains(pr.AssignedReviewers, userId) {
		return nil, "", ErrUserNotAReviewer
	}

	teamName, err := s.GetTeamNameByUserId(ctx, tx, pr.AuthorId)
	if err != nil {
		return nil, "", err
	}

	candidate, err := s.GetReviewers(
		ctx,
		tx,
		teamName,
		append(pr.AssignedReviewers, pr.AuthorId),
		1,
	)
	if err != nil {
		return nil, "", err
	} else if len(candidate) == 0 {
		return nil, "", ErrNoCandidate
	}

	pr.AssignedReviewers[slices.Index(pr.AssignedReviewers, userId)] = candidate[0]
	sql := `UPDATE pull_requests
	SET
		assigned_reviewers = $1
	WHERE pull_request_id = $2`
	if _, err = tx.Exec(ctx, sql, pr.AssignedReviewers, pullRequestId); err != nil {
		return nil, "", fmt.Errorf("%v failed to execute update: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, "", fmt.Errorf("%v failed to commit transaction: %w", op, err)
	}

	return pr, candidate[0], nil
}

func (s *Storage) CreatePullRequest(
	ctx context.Context,
	req api.PostPullRequestCreateJSONBody,
) (*api.PullRequest, error) {
	const op = "postgres.CreatePullRequest"
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v failed to begin transaction: %w", op, err)
	}
	defer Rollback(ctx, tx)

	authorTeam, err := s.GetTeamNameByUserId(ctx, tx, req.AuthorId)
	if err != nil {
		return nil, err
	}

	if ok, err := s.IsPullRequestExists(ctx, tx, req.PullRequestId); err != nil {
		return nil, err
	} else if ok {
		return nil, ErrPullRequestExists
	}

	reviewers, err := s.GetReviewers(ctx, tx, authorTeam, []string{req.AuthorId}, 2)
	if err != nil {
		return nil, err
	}

	sql := `INSERT INTO pull_requests 
	(pull_request_id, pull_request_name, author_id, assigned_reviewers)
	VALUES ($1, $2, $3, $4)
	RETURNING pull_request_id, pull_request_name, author_id, assigned_reviewers, status, createdAt, mergedAt`
	var pr api.PullRequest
	err = tx.QueryRow(
		ctx,
		sql,
		req.PullRequestId,
		req.PullRequestName,
		req.AuthorId,
		reviewers,
	).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.AssignedReviewers,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%v failed to query row: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%v failed to commit transaction: %w", op, err)
	}

	return &pr, nil
}

func (s *Storage) GetReviewers(
	ctx context.Context,
	tx pgx.Tx,
	teamName string,
	tabu []string,
	count int,
) ([]string, error) {
	const op = "postgres.GetReviewers"
	sql := `SELECT user_id FROM users
	WHERE team_name = $1 AND 
		user_id != ALL($2) AND 
		is_active = true
	ORDER BY RANDOM()
	LIMIT $3`
	rows, err := tx.Query(ctx, sql, teamName, tabu, count)
	if err != nil {
		return nil, fmt.Errorf("%v failed to query: %w", op, err)
	}

	reviewers, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var r string
		return r, row.Scan(&r)
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to collect rows: %w", op, err)
	}
	return reviewers, nil
}

func (s *Storage) IsPullRequestExists(ctx context.Context, tx pgx.Tx, prId string) (bool, error) {
	sql := "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)"
	var ok bool
	if err := tx.QueryRow(ctx, sql, prId).Scan(&ok); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ok, fmt.Errorf("postgres.IsPullRequestExists failed to query row: %w", err)
	}
	return ok, nil
}

func (s *Storage) GetPullRequest(ctx context.Context, tx pgx.Tx, prId string) (*api.PullRequest, error) {
	sql := `SELECT 
		pull_request_id,
		pull_request_name, 
		author_id,
		assigned_reviewers,
		status,
		createdAt,
		mergedAt 
	FROM pull_requests
	WHERE pull_request_id = $1`
	var pr api.PullRequest
	err := tx.QueryRow(ctx, sql, prId).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.AssignedReviewers,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("postgres.GetPullRequests failed to query row: %w", err)
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPullRequestNotFound
	}
	return &pr, nil
}
