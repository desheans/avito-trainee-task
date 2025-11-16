package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"avito-trainee-task/internal/api"
	"avito-trainee-task/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

// PostPullRequestCreate implements api.ServerInterface.
func (h *Handler) PostPullRequestCreate(c echo.Context) error {
	ctx := c.Request().Context()

	var req api.PostPullRequestCreateJSONBody
	if err := c.Bind(&req); err != nil {
		slog.ErrorContext(
			ctx, "failed to bind request",
			"error", err)
		return echo.ErrBadRequest
	}

	pr, err := h.s.CreatePullRequest(ctx, req)
	switch {
	case errors.Is(err, postgres.ErrUserNotFound):
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "Author/Team not foun",
		))
	case errors.Is(err, postgres.ErrPullRequestExists):
		return c.JSON(http.StatusConflict, NewError(
			api.PREXISTS, "PR id already exists",
		))
	case err != nil:
		slog.ErrorContext(ctx, "failed to create pull request", "error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, &struct {
		PR api.PullRequest `json:"pr"`
	}{
		PR: *pr,
	})
}

// PostPullRequestMerge implements api.ServerInterface.
func (h *Handler) PostPullRequestMerge(c echo.Context) error {
	ctx := c.Request().Context()

	var req api.PostPullRequestMergeJSONBody
	if err := c.Bind(&req); err != nil {
		slog.ErrorContext(ctx, "failed to bind request",
			"error", err)
		return echo.ErrBadRequest
	}

	pr, err := h.s.Merge(ctx, req.PullRequestId)
	if errors.Is(err, postgres.ErrPullRequestNotFound) {
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "Pull request not found",
		))
	} else if err != nil {
		slog.ErrorContext(ctx, "failed to merge pull request",
			"error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, &struct {
		PR api.PullRequest `json:"pr"`
	}{
		PR: *pr,
	})
}

// PostPullRequestReassign implements api.ServerInterface.
func (h *Handler) PostPullRequestReassign(c echo.Context) error {
	ctx := c.Request().Context()

	var req api.PostPullRequestReassignJSONBody
	if err := c.Bind(&req); err != nil {
		slog.ErrorContext(ctx, "failed to bind request",
			"error", err)
		return echo.ErrBadRequest
	}

	pr, new, err := h.s.Reassign(
		c.Request().Context(),
		req.PullRequestId,
		req.OldUserId,
	)

	switch {
	case errors.Is(err, postgres.ErrReassignMergedPullRequest):
		return c.JSON(http.StatusConflict, NewError(
			api.PRMERGED, "Cannot reassign on mergedd PR",
		))
	case errors.Is(err, postgres.ErrUserNotAReviewer):
		return c.JSON(http.StatusConflict, NewError(
			api.NOTASSIGNED, "Reviewer is not assigned to this PR",
		))
	case errors.Is(err, postgres.ErrNoCandidate):
		return c.JSON(http.StatusConflict, NewError(
			api.NOCANDIDATE, "No active replacement candidate in team",
		))
	case errors.Is(err, postgres.ErrUserNotFound):
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "User not found",
		))
	case errors.Is(err, postgres.ErrPullRequestNotFound):
		return c.JSON(http.StatusNotFound, NewError(
			api.NOTFOUND, "Pull request not found",
		))
	case err != nil:
		slog.ErrorContext(ctx, "failed to merge reassign request",
			"error", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, &struct {
		PR         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{
		PR:         *pr,
		ReplacedBy: new,
	})
}
