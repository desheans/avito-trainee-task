package storage

import (
	"testing"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/stretchr/testify/require"
)

func TestGetTeam(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('user1', 'alice', 'backend', true),
			('user2', 'bob', 'backend', true),
			('user3', 'charlie', 'frontend', true)
		`)
	require.NoError(t, err)

	team, err := storage.GetTeam(ctx, "backend")
	require.NoError(t, err)

	require.Equal(t, "backend", team.TeamName)
	require.Len(t, team.Members, 2)

	memberIDs := make([]string, len(team.Members))
	for i, m := range team.Members {
		memberIDs[i] = m.UserId
	}
	require.ElementsMatch(t, []string{"user1", "user2"}, memberIDs)

	var alice api.TeamMember
	for _, m := range team.Members {
		if m.UserId == "user1" {
			alice = m
			break
		}
	}
	require.Equal(t, "alice", alice.Username)
	require.True(t, alice.IsActive)
}

func TestGetNonExistentTeam(t *testing.T) {
	_, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	team, err := storage.GetTeam(ctx, "NONEXISTENT")
	require.Nil(t, team)
	require.ErrorIs(t, err, postgres.ErrTeamNotFound)
}

func TestAddTeam(t *testing.T) {
	_, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	expected := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{UserId: "user1", Username: "alice", IsActive: true},
			{UserId: "user2", Username: "bob", IsActive: true},
		},
	}

	result, err := storage.AddTeam(ctx, expected)
	require.NoError(t, err)

	require.Equal(t, expected.TeamName, result.TeamName)
	require.Len(t, result.Members, len(expected.Members))

	actual, err := storage.GetTeam(ctx, expected.TeamName)
	require.NoError(t, err)
	require.Equal(t, expected.TeamName, actual.TeamName)
	require.Len(t, actual.Members, len(expected.Members))
}

func TestAddTeamUpdateUser(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('user1', 'alice', 'frontend', true)
		`)
	require.NoError(t, err)

	expected := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{UserId: "user1", Username: "alice", IsActive: true},
		},
	}

	_, err = storage.AddTeam(ctx, expected)
	require.NoError(t, err)

	actual, err := storage.GetTeam(ctx, expected.TeamName)
	require.NoError(t, err)
	require.Equal(t, expected.TeamName, actual.TeamName)
	require.Equal(t, "user1", actual.Members[0].UserId)
}

func TestAddTeamAlreadyExists(t *testing.T) {
	_, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	team1 := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{UserId: "user1", Username: "alice", IsActive: true},
		},
	}
	_, err := storage.AddTeam(ctx, team1)
	require.NoError(t, err)

	team2 := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{UserId: "user2", Username: "bob", IsActive: true},
		},
	}
	result, err := storage.AddTeam(ctx, team2)
	require.Nil(t, result)
	require.ErrorIs(t, err, postgres.ErrTeamExists)
}

func TestIsTeamExists(t *testing.T) {
	tx, storage, cleanup := setupTestStorage(t)
	defer cleanup()

	_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('user1', 'alice', 'frontend', true)
		`)
	require.NoError(t, err)

	ok, err := storage.IsTeamExists(ctx, tx, "frontend")
	require.NoError(t, err)
	require.True(t, ok)

	expected := api.Team{
		TeamName: "backend",
		Members: []api.TeamMember{
			{UserId: "user1", Username: "alice", IsActive: true},
		},
	}
	_, err = storage.AddTeam(ctx, expected)
	require.NoError(t, err)

	ok, err = storage.IsTeamExists(ctx, tx, "frontend")
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = storage.IsTeamExists(ctx, tx, expected.TeamName)
	require.NoError(t, err)
	require.True(t, ok)
}
