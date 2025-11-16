package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

// GetTeamGet implements api.ServerInterface.
func (h *Handler) GetTeamGet(c echo.Context, params api.GetTeamGetParams) error {
	ctx := c.Request().Context()

	team, err := h.s.GetTeam(ctx, params.TeamName)
	if errors.Is(err, postgres.ErrTeamNotFound) {
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "Team not found",
		))
	} else if err != nil {
		slog.ErrorContext(
			ctx,
			"failed to get team",
			"team_name", params.TeamName,
			"error", err,
		)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, team)
}

// PostTeamAdd implements api.ServerInterface.
func (h *Handler) PostTeamAdd(c echo.Context) error {
	ctx := c.Request().Context()

	var team api.Team
	if err := c.Bind(&team); err != nil {
		return echo.ErrBadRequest
	}

	addedTeam, err := h.s.AddTeam(ctx, team)
	if errors.Is(err, postgres.ErrTeamExists) {
		return c.JSON(http.StatusBadRequest, NewError(
			api.TEAMEXISTS, "team_name already exists",
		))
	} else if err != nil {
		slog.ErrorContext(ctx, "failed to add team", "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, &struct {
		Team api.Team `json:"team"`
	}{
		Team: *addedTeam,
	})
}
