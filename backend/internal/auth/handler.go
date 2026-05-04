package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tokenTTL = 90 * 24 * time.Hour

type Service struct {
	JWTSecret string
	GitHub    *GitHubOAuth
	Logger    *zap.Logger
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/auth")

	// Dev-only: выдаём токен без OAuth для локальной разработки.
	// В production отключается, если EOP_ENV=production.
	g.Post("/dev-token", func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			userID = uuid.NewString()
		}
		tok, err := IssueJWT(s.JWTSecret, userID, "dev@local", "dev", tokenTTL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID})
	})

	g.Get("/github/login", func(c *fiber.Ctx) error {
		state := randomState()
		c.Cookie(&fiber.Cookie{
			Name:     "eop_oauth_state",
			Value:    state,
			HTTPOnly: true,
			SameSite: "Lax",
			MaxAge:   600,
		})
		return c.Redirect(s.GitHub.AuthCodeURL(state))
	})

	g.Get("/github/callback", func(c *fiber.Ctx) error {
		got := c.Query("state")
		want := c.Cookies("eop_oauth_state")
		if got == "" || got != want {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "state mismatch"})
		}
		code := c.Query("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing code"})
		}
		user, err := s.GitHub.Exchange(c.Context(), code)
		if err != nil {
			s.Logger.Warn("github exchange failed", zap.Error(err))
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "oauth exchange failed"})
		}
		// Phase 2: upsert в Postgres users и связать device.
		// Phase 1: сразу выдаём JWT с github_login как сабом.
		userID := "gh:" + user.Login
		tok, err := IssueJWT(s.JWTSecret, userID, user.Email, "github", tokenTTL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID, "github_login": user.Login})
	})
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
