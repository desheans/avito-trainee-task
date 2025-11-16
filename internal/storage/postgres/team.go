package postgres

import (
	"context"
	"errors"
	"fmt"

	"avito-trainee-task/internal/api"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	const op = "postgres.GetTeam"
	rows, err := s.db.Query(ctx,
		`SELECT user_id, username, is_active FROM users 
		WHERE team_name = $1`,
		teamName)
	if err != nil {
		return nil, fmt.Errorf("%v failed to query: %w", op, err)
	}

	members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (api.TeamMember, error) {
		var m api.TeamMember
		return m, row.Scan(
			&m.UserId,
			&m.Username,
			&m.IsActive,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("%v failed to collect rows: %w", op, err)
	}

	if len(members) == 0 {
		return nil, ErrTeamNotFound
	}

	return &api.Team{
		Members:  members,
		TeamName: teamName,
	}, nil
}

func (s *Storage) AddTeam(ctx context.Context, team api.Team) (*api.Team, error) {
	const op = "postgres.AddTeam"
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v failed to begin transaction: %w", op, err)
	}
	defer Rollback(ctx, tx)

	exists, err := s.IsTeamExists(ctx, tx, team.TeamName)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, ErrTeamExists
	}

	sql := `INSERT INTO users (user_id, username, team_name, is_active)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id)
	DO UPDATE SET 
        team_name = EXCLUDED.team_name
	RETURNING user_id, username, is_active`
	var members []api.TeamMember
	for _, m := range team.Members {
		var member api.TeamMember
		err := tx.QueryRow(ctx, sql, m.UserId, m.Username, team.TeamName, m.IsActive).Scan(
			&member.UserId,
			&member.Username,
			&member.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("%v failed to query row: %w", op, err)
		}
		members = append(members, member)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%v failed to commit transaction: %w", op, err)
	}

	return &api.Team{
		TeamName: team.TeamName,
		Members:  members,
	}, nil
}

func (s *Storage) IsTeamExists(ctx context.Context, tx pgx.Tx, teamName string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE team_name = $1)`
	var ok bool
	err := tx.QueryRow(ctx, sql, teamName).Scan(&ok)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ok, fmt.Errorf("postgres.IsTeamExists failed to query row: %w", err)
	}
	return ok, nil
}
