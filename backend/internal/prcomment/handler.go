package prcomment

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth"
	"github.com/eye-of-providence/backend/internal/httperr"
)

type Service struct {
	Pool         *pgxpool.Pool
	JWTSecret    string
	Logger       *zap.Logger
	HTTPClient   HTTPClient
	DashboardURL string
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/integrations", auth.Middleware(s.JWTSecret, s.Pool))
	g.Post("/pr-comment", postHandler(s))
}

type request struct {
	Provider      string   `json:"provider"`
	Host          string   `json:"host"`
	Repo          string   `json:"repo"`
	PRNumber      int      `json:"pr_number"`
	SHAs          []string `json:"shas"`
	ProviderToken string   `json:"provider_token"`
}

func postHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		var req request
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if err := validate(req); err != nil {
			return httperr.BadRequest(c, "invalid_request", err.Error())
		}

		comment, agg, err := (&CommentBody{Pool: s.Pool, Base: s.DashboardURL}).Markdown(c.Context(), uid, req.SHAs)
		if err != nil {
			s.Logger.Error("pr-comment aggregate failed", zap.Error(err))
			return httperr.Internal(c)
		}

		if err := postComment(c.Context(), s.HTTPClient, req, comment); err != nil {
			s.Logger.Warn("pr-comment post failed",
				zap.String("provider", req.Provider),
				zap.Error(err))
			var pe *PostError
			if errors.As(err, &pe) {
				return httperr.Send(c, httperr.ProblemDetails{
					Status: fiber.StatusBadGateway,
					Code:   "provider_rejected",
					Detail: "provider rejected comment",
					Extensions: map[string]any{
						"provider_status": pe.Status,
						"provider_body":   pe.Body,
					},
				})
			}
			return httperr.BadGateway(c, "provider_request_failed", "provider request failed")
		}

		return c.JSON(fiber.Map{
			"posted":     true,
			"aggregate":  agg,
			"comment_md": comment,
		})
	}
}

func postComment(ctx context.Context, hc HTTPClient, req request, body string) error {
	switch req.Provider {
	case "github":
		return PostGitHub(ctx, hc, req.Host, req.Repo, req.PRNumber, req.ProviderToken, body)
	case "gitlab":
		return PostGitLab(ctx, hc, req.Host, req.Repo, req.PRNumber, req.ProviderToken, body)
	}
	return errors.New("unknown provider")
}

func validate(r request) error {
	switch r.Provider {
	case "github", "gitlab":
	default:
		return errors.New("provider must be 'github' or 'gitlab'")
	}
	if strings.TrimSpace(r.Repo) == "" || !strings.Contains(r.Repo, "/") {
		return errors.New("repo must be 'owner/name'")
	}
	if r.PRNumber <= 0 {
		return errors.New("pr_number must be positive")
	}
	if len(r.SHAs) == 0 {
		return errors.New("shas required")
	}
	if len(r.SHAs) > 500 {
		return errors.New("too many shas (max 500)")
	}
	if strings.TrimSpace(r.ProviderToken) == "" {
		return errors.New("provider_token required")
	}
	return nil
}
