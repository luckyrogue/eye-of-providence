package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const tokenTTL = 90 * 24 * time.Hour

type Service struct {
	JWTSecret      string
	GitHub         *GitHubOAuth
	Logger         *zap.Logger
	Users          *UsersPG
	EnableDevToken bool // false в production — роут /dev-token не регистрируется
}

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/auth")

	if s.EnableDevToken {
		registerDevToken(g, s)
	}

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
		// Детерминированный UUID на основе github user.id — позволяет
		// ClickHouse user_id оставаться UUID-типизированным.
		userUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(fmt.Sprintf("github:%d", user.ID)))
		email := user.Email
		if email == "" {
			email = fmt.Sprintf("github-%s@local.eop", user.Login)
		}
		if err := s.Users.Upsert(c.Context(), userUUID, email, user.Login); err != nil {
			s.Logger.Warn("github user upsert failed", zap.Error(err))
		}
		tok, err := IssueJWT(s.JWTSecret, userUUID.String(), email, "github", tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "auth failed"})
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userUUID.String(), "github_login": user.Login})
	})
}

// registerDevToken — выдаёт JWT без OAuth (для локальной разработки).
// Регистрируется только когда EnableDevToken=true (по умолчанию выключен в production).
func registerDevToken(g fiber.Router, s Service) {
	g.Post("/dev-token", func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		var userID uuid.UUID
		if userIDStr == "" {
			userID = uuid.New()
		} else {
			parsed, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id must be uuid"})
			}
			userID = parsed
		}
		email := fmt.Sprintf("dev-%s@local.eop", userID.String()[:8])
		if err := s.Users.Upsert(c.Context(), userID, email, ""); err != nil {
			s.Logger.Warn("user upsert failed (continuing with token)", zap.Error(err))
		}
		tok, err := IssueJWT(s.JWTSecret, userID.String(), email, "dev", tokenTTL)
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "auth failed"})
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID.String()})
	})
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
