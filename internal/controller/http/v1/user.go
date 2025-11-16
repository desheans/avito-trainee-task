package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

// PostUsersSetIsActive implements api.ServerInterface.
func (h *Handler) PostUsersSetIsActive(c echo.Context) error {
	ctx := c.Request().Context()

	var req api.PostUsersSetIsActiveJSONBody
	if err := c.Bind(&req); err != nil {
		slog.ErrorContext(ctx, "failed to bind request", "error", err)
		return echo.ErrBadRequest
	}

	user, err := h.s.SetIsActive(
		ctx,
		req.UserId,
		req.IsActive,
	)

	if errors.Is(err, postgres.ErrUserNotFound) {
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "User not found",
		))
	} else if err != nil {
		slog.ErrorContext(ctx, "failed to set isActive", "user_id", user.UserId, "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, &struct {
		User api.User `json:"user"`
	}{
		User: *user,
	})
}

// GetUsersGetReview implements api.ServerInterface.
func (h *Handler) GetUsersGetReview(c echo.Context, params api.GetUsersGetReviewParams) error {
	ctx := c.Request().Context()
	prs, err := h.s.GetReview(ctx, params.UserId)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "User not found",
		))
	} else if err != nil {
		slog.ErrorContext(ctx, "filed to get review", "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, &struct {
		UserId       string                  `json:"user_id"`
		PullRequests []*api.PullRequestShort `json:"pull_requests"`
	}{
		UserId:       params.UserId,
		PullRequests: prs,
	})
}

// GetUsersStats implements api.ServerInterface.
func (h *Handler) GetUsersStats(c echo.Context) error {
	ctx := c.Request().Context()
	stats, err := h.s.GetUsersStats(ctx)
	if err != nil {
		slog.ErrorContext(
			ctx, "failed to get users status",
			"error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, stats)
}
