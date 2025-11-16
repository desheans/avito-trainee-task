package v1

import (
	"context"
	"log/slog"
	"time"

	"avito-trainee-task/internal/api"

	"github.com/labstack/echo/v4"
)

type Storage interface {
	SetIsActive(ctx context.Context, UserId string, isActive bool) (*api.User, error)
	GetReview(ctx context.Context, userId string) ([]*api.PullRequestShort, error)
	GetUsersStats(ctx context.Context) (*api.AssignmentCountStat, error)

	GetTeam(ctx context.Context, teamName string) (*api.Team, error)
	AddTeam(ctx context.Context, team api.Team) (*api.Team, error)

	Merge(ctx context.Context, pullRequestId string) (*api.PullRequest, error)
	Reassign(ctx context.Context, pullRequestId, userId string) (*api.PullRequest, string, error)
	CreatePullRequest(ctx context.Context, req api.PostPullRequestCreateJSONBody) (*api.PullRequest, error)
}

type Handler struct {
	s Storage
}

func NewHandler(s Storage) *Handler {
	return &Handler{
		s: s,
	}
}

func (h *Handler) RegisterRoutes(e api.EchoRouter) {
	api.RegisterHandlersWithBaseURL(e, h, "")
}

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		start := time.Now()
		err := next(c)
		duration := time.Since(start)

		if duration > 500*time.Millisecond {
			slog.WarnContext(
				ctx, "slow request",
				"method", c.Request().Method,
				"path", c.Path(),
				"duration_ms", duration.Milliseconds(),
			)
		}

		if err != nil {
			slog.ErrorContext(
				ctx, "request failed",
				"method", c.Request().Method,
				"path", c.Path(),
				"error", err,
			)
		}

		return err
	}
}

type ErrorResponseBody struct {
	Code    api.ErrorResponseErrorCode `json:"code"`
	Message string                     `json:"message"`
}

func NewError(code api.ErrorResponseErrorCode, message string) *api.ErrorResponse {
	return &api.ErrorResponse{
		Error: ErrorResponseBody{
			Code:    code,
			Message: message,
		},
	}
}
