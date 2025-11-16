package storage

import (
	"testing"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/stretchr/testify/require"
)

func TestSetIsActive(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	u := api.User{
		IsActive: false,
		TeamName: "android",
		UserId:   "USR006",
		Username: "anna_smirnova",
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)`, u.UserId, u.Username, u.TeamName, u.IsActive)
	require.NoError(t, err)

	new, err := storage.SetIsActive(ctx, u.UserId, !u.IsActive)
	require.NoError(t, err)
	require.Equal(t, u.IsActive, !new.IsActive)

	var isActive bool
	err = tx.QueryRow(ctx,
		`SELECT is_active FROM users WHERE user_id = $1`,
		u.UserId).Scan(&isActive)
	require.NoError(t, err)
	require.Equal(t, !u.IsActive, isActive)
}

func TestSetIsActiveNonExistUser(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ('USR006', 'anna_smirnova', 'android', FALSE)`)
	require.NoError(t, err)

	_, err = storage.SetIsActive(ctx, "NONEXISTUSERID", true)
	require.ErrorIs(t, err, postgres.ErrUserNotFound)
}

func TestGetReview(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('author1', 'author1', 'backend', true),
			('reviewer1', 'reviewer1', 'backend', true),
			('reviewer2', 'reviewer2', 'backend', true)
		`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `
			INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, assigned_reviewers, status) VALUES
			('pr1', 'PR 1', 'author1', '{"reviewer1","reviewer2"}', 'OPEN'),
			('pr2', 'PR 2', 'author1', '{"reviewer1"}', 'MERGED'),
			('pr3', 'PR 3', 'author1', '{"reviewer2"}', 'OPEN')
		`)
	require.NoError(t, err)

	prs, err := storage.GetReview(ctx, "reviewer1")
	require.NoError(t, err)

	require.Len(t, prs, 2)

	prIDs := make([]string, len(prs))
	for i, pr := range prs {
		prIDs[i] = pr.PullRequestId
	}
	require.ElementsMatch(t, []string{"pr1", "pr2"}, prIDs)

	var foundPR *api.PullRequestShort
	for _, pr := range prs {
		if pr.PullRequestId == "pr1" {
			foundPR = pr
			break
		}
	}
	require.NotNil(t, foundPR)
	require.Equal(t, "PR 1", foundPR.PullRequestName)
	require.Equal(t, "author1", foundPR.AuthorId)
	require.Equal(t, api.PullRequestShortStatusOPEN, foundPR.Status)
}

func TestGetTeamByUserId(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	u := api.User{
		IsActive: false,
		TeamName: "android",
		UserId:   "USR006",
		Username: "anna_smirnova",
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)`, u.UserId, u.Username, u.TeamName, u.IsActive)
	require.NoError(t, err)

	teamName, err := storage.GetTeamNameByUserId(ctx, tx, u.UserId)
	require.NoError(t, err)
	require.Equal(t, u.TeamName, teamName)
}

func TestGetTeamByUserIdNonExistUser(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := storage.GetTeamNameByUserId(ctx, tx, "FFFFFF")
	require.ErrorIs(t, err, postgres.ErrUserNotFound)
}
