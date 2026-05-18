package auth

import (
	"encoding/base64"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
)

const (
	tokenTTL = 14 * 24 * time.Hour

	handoffCookieName = "eop_session_handoff"
	handoffTTL        = 30 * time.Second
)

type Service struct {
	JWTSecret string

	Providers      map[string]OAuthProvider
	GitHub         *GitHubOAuth
	Logger         *zap.Logger
	Users          *UsersPG
	Pool           *pgxpool.Pool
	WebAuthn       *WebAuthnService
	PublicURL      string
	SecureCookies  bool
	EnableDevToken bool
}

func (s Service) ProvidersList() []string {
	out := make([]string, 0, len(s.Providers))
	for name := range s.Providers {
		out = append(out, name)
	}

	order := []string{"github", "google", "apple"}
	sorted := make([]string, 0, len(out))
	for _, k := range order {
		for _, n := range out {
			if n == k {
				sorted = append(sorted, n)
				break
			}
		}
	}

	for _, n := range out {
		found := false
		for _, k := range order {
			if n == k {
				found = true
				break
			}
		}
		if !found {
			sorted = append(sorted, n)
		}
	}
	return sorted
}

func setHandoffCookie(c *fiber.Ctx, s Service, tok string) {
	c.Cookie(&fiber.Cookie{
		Name:     handoffCookieName,
		Value:    tok,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   s.SecureCookies,
		MaxAge:   int(handoffTTL.Seconds()),
	})
}

func (s Service) internalErr(c *fiber.Ctx, err error) error {
	rid, _ := c.Locals("requestid").(string)
	s.Logger.Error("internal error",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("rid", rid),
		zap.Error(err),
	)
	return httperr.Internal(c)
}

func base64URLEncode(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}
