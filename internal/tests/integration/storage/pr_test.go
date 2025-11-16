package storage

import (
	"testing"
	"time"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
		INSERT INTO users 
			(user_id, username, team_name, is_active) 
			VALUES ('author1', 'author', 'backend', true);
		INSERT INTO pull_requests 
			(pull_request_id, pull_request_name, author_id, assigned_reviewers, status) 
			VALUES ('pr1', 'Test PR', 'author1', '{"rev1"}', 'OPEN')`)
	require.NoError(t, err)

	result, err := storage.Merge(ctx, "pr1")
	require.NoError(t, err)

	require.Equal(t, "pr1", result.PullRequestId)
	require.Equal(t, api.PullRequestStatusMERGED, result.Status)
	require.NotNil(t, result.MergedAt)

	var status api.PullRequestStatus
	var mergedAt *time.Time
	err = tx.QueryRow(ctx,
		"SELECT status, mergedAt FROM pull_requests WHERE pull_request_id = 'pr1'").Scan(&status, &mergedAt)
	require.NoError(t, err)
	require.Equal(t, api.PullRequestStatusMERGED, status)
	require.NotNil(t, mergedAt)
}

func TestMergeNonExistentPR(t *testing.T) {
	_, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	pr, err := storage.Merge(ctx, "NONEXISTENT")
	require.Nil(t, pr)
	require.ErrorIs(t, err, postgres.ErrPullRequestNotFound)
}

func TestMergeMergedPR(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active) VALUES
		('author1', 'author', 'backend', true)`)
	require.NoError(t, err)

	mergedTime := time.Now().Add(-time.Hour)
	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests 
		(pull_request_id, pull_request_name, author_id, assigned_reviewers, status, mergedAt) 
		VALUES('pr1', 'Test PR', 'author1', '{"rev1"}', 'MERGED', $1)`,
		mergedTime)
	require.NoError(t, err)

	mergedPR, err := storage.Merge(ctx, "pr1")
	require.NoError(t, err)

	require.True(t, mergedPR.MergedAt.After(mergedTime))
}

func TestIsPRExistsTrue(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
		INSERT INTO users
			(user_id, username, team_name, is_active)
			VALUES ('author1', 'author', 'backend', 'true');
		INSERT INTO pull_requests 
			(pull_request_id, pull_request_name, author_id, status) 
			VALUES ('pr1', 'Test PR', 'author1', 'OPEN')
		`)
	require.NoError(t, err)

	exists, err := storage.IsPullRequestExists(ctx, tx, "pr1")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestIsPRExistsFalse(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	exists, err := storage.IsPullRequestExists(ctx, tx, "NONEXISTENT")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestGetPR(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active) VALUES
		('author1', 'author', 'backend', true)`)
	require.NoError(t, err)

	now := time.Now()
	exp := api.PullRequest{
		AssignedReviewers: []string{"rev1", "rev2"},
		AuthorId:          "author1",
		CreatedAt:         &now,
		PullRequestId:     "pr1",
		PullRequestName:   "Test PR",
		Status:            api.PullRequestStatusOPEN,
	}
	_, err = tx.Exec(
		ctx,
		`
			INSERT INTO pull_requests 
			(pull_request_id, pull_request_name, author_id, assigned_reviewers, status, createdAt) 
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
		exp.PullRequestId,
		exp.PullRequestName,
		exp.AuthorId,
		exp.AssignedReviewers,
		exp.Status,
		exp.CreatedAt,
	)
	require.NoError(t, err)

	act, err := storage.GetPullRequest(ctx, tx, exp.PullRequestId)
	require.NoError(t, err)

	require.Equal(t, exp.PullRequestId, act.PullRequestId)
	require.Equal(t, exp.PullRequestName, act.PullRequestName)
	require.Equal(t, exp.AuthorId, act.AuthorId)
	require.Equal(t, exp.AssignedReviewers, act.AssignedReviewers)
	require.Equal(t, exp.Status, act.Status)
	require.Nil(t, act.MergedAt)
}

func TestGetNonExistentPR(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	pr, err := storage.GetPullRequest(ctx, tx, "NONEXISTENT")
	require.Nil(t, pr)
	require.ErrorIs(t, err, postgres.ErrPullRequestNotFound)
}

func TestCreatePR(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'alice', 'backend', true),
			('reviewer1', 'bob', 'backend', true),
			('reviewer2', 'charlie', 'backend', true)
		`)
	require.NoError(t, err)

	exp := api.PostPullRequestCreateJSONBody{
		PullRequestId:   "pr1",
		PullRequestName: "Test PR",
		AuthorId:        "author1",
	}

	act, err := storage.CreatePullRequest(ctx, exp)
	require.NoError(t, err)

	require.Equal(t, exp.PullRequestId, act.PullRequestId)
	require.Equal(t, exp.PullRequestName, act.PullRequestName)
	require.Equal(t, exp.AuthorId, act.AuthorId)
	require.Equal(t, api.PullRequestStatusOPEN, act.Status)
	require.Len(t, act.AssignedReviewers, 2)
	require.NotNil(t, act.CreatedAt)
	require.Nil(t, act.MergedAt)

	var pr api.PullRequest
	err = tx.QueryRow(ctx,
		"SELECT * FROM pull_requests WHERE pull_request_id = 'pr1'").Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.AssignedReviewers,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	require.NoError(t, err)
	require.Equal(t, act, &pr)
}

func TestReassign(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'alice', 'backend', true),
			('reviewer1', 'bob', 'backend', true),
			('reviewer2', 'charlie', 'backend', true),
			('reviewer3', 'dave', 'backend', true)
		`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `
			INSERT INTO pull_requests 
			(pull_request_id, pull_request_name, author_id, assigned_reviewers, status) 
			VALUES ('pr1', 'Test PR', 'author1', '{"reviewer1","reviewer2"}', 'OPEN')
		`)
	require.NoError(t, err)

	actual, newRev, err := storage.Reassign(ctx, "pr1", "reviewer1")
	require.NoError(t, err)

	require.Equal(t, "pr1", actual.PullRequestId)
	require.Contains(t, actual.AssignedReviewers, newRev)
	require.NotContains(t, actual.AssignedReviewers, "reviewer1")
	require.Len(t, actual.AssignedReviewers, 2)
}

func TestReassignNonExistent(t *testing.T) {
	_, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	pr, newReviewer, err := storage.Reassign(ctx, "NONEXISTENT", "user1")
	require.Nil(t, pr)
	require.Empty(t, newReviewer)
	require.ErrorIs(t, err, postgres.ErrPullRequestNotFound)
}

func TestReassignMerged(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'alice', 'backend', true),
			('reviewer1', 'bob', 'backend', true)
		`)
	require.NoError(t, err)

	_, err = storage.CreatePullRequest(ctx, api.PostPullRequestCreateJSONBody{
		AuthorId:        "author1",
		PullRequestId:   "pr1",
		PullRequestName: "Test PR",
	})
	require.NoError(t, err)

	_, err = storage.Merge(ctx, "pr1")
	require.NoError(t, err)

	pr, newReviewer, err := storage.Reassign(ctx, "pr1", "reviewer1")
	require.Nil(t, pr)
	require.Empty(t, newReviewer)
	require.ErrorIs(t, err, postgres.ErrReassignMergedPullRequest)
}

func TestReassignUserNotAReviewer(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'alice', 'backend', true),
			('reviewer1', 'bob', 'backend', true)
		`)
	require.NoError(t, err)

	_, err = storage.CreatePullRequest(ctx, api.PostPullRequestCreateJSONBody{
		AuthorId:        "author1",
		PullRequestId:   "pr1",
		PullRequestName: "Test PR",
	})
	require.NoError(t, err)

	pr, newRev, err := storage.Reassign(ctx, "pr1", "NONEXISTENTUSER")
	require.Nil(t, pr)
	require.Empty(t, newRev)
	require.ErrorIs(t, err, postgres.ErrUserNotAReviewer)
}

func TestReassignNoCandidate(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'alice', 'backend', true),
			('reviewer1', 'bob', 'backend', true)
		`)
	require.NoError(t, err)

	_, err = storage.CreatePullRequest(ctx, api.PostPullRequestCreateJSONBody{
		AuthorId:        "author1",
		PullRequestId:   "pr1",
		PullRequestName: "Test PR",
	})
	require.NoError(t, err)

	pr, newRev, err := storage.Reassign(ctx, "pr1", "reviewer1")
	require.Nil(t, pr)
	require.Empty(t, newRev)
	require.ErrorIs(t, err, postgres.ErrNoCandidate)
}
