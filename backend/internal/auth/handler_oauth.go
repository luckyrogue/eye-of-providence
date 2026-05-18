package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/auth/oauthflowapp"
	"github.com/eye-of-providence/backend/internal/auth/sessionapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func RegisterRoutes(app *fiber.App, s Service) {
	g := app.Group("/v1/auth")

	if s.EnableDevToken {
		registerDevToken(g, s)
	}

	for name, p := range s.Providers {
		registerOAuthProvider(g, s, name, p)
	}

	if _, has := s.Providers["github"]; !has && s.GitHub != nil {
		registerOAuthProvider(g, s, "github", s.GitHub)
	}

	g.Get("/session-handoff-internal", handleSessionHandoff(s))

	if s.WebAuthn != nil {
		g.Post("/webauthn/login/begin", handleWebAuthnLoginBegin(s))
		g.Post("/webauthn/login/finish", handleWebAuthnLoginFinish(s))

		mw := Middleware(s.JWTSecret, s.Pool)
		g.Post("/webauthn/register/begin", mw, handleWebAuthnRegisterBegin(s))
		g.Post("/webauthn/register/finish", mw, handleWebAuthnRegisterFinish(s))
	}
}

func RegisterSessionHandoffRoute(app *fiber.App, s Service) {
	app.Get("/v1/me/session-handoff", handleSessionHandoff(s))
}

func registerOAuthProvider(g fiber.Router, s Service, name string, p OAuthProvider) {
	flow := newOAuthFlowApp(s)
	g.Get("/"+name+"/login", func(c *fiber.Ctx) error {
		nonce := flow.RandomState()
		returnTo := c.Query("return_to")
		c.Cookie(&fiber.Cookie{
			Name:     "eop_oauth_state",
			Value:    flow.StateCookieValue(nonce, returnTo),
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
			MaxAge:   600,
		})
		return c.Redirect(p.AuthCodeURL(nonce))
	})

	g.Get("/"+name+"/callback", func(c *fiber.Ctx) error {
		res, err := flow.CompleteCallback(c.Context(), name, oauthProviderAdapter{p: p}, oauthflowapp.CallbackInput{
			GotState:          c.Query("state"),
			StoredStateCookie: c.Cookies("eop_oauth_state"),
			Code:              c.Query("code"),
			OAuthError:        c.Query("error"),
		})
		if err != nil {
			switch {
			case errors.Is(err, oauthflowapp.ErrStateMismatch):
				return httperr.BadRequest(c, "state_mismatch", "state mismatch")
			case errors.Is(err, oauthflowapp.ErrUserDenied):
				return httperr.BadRequest(c, "user_denied", "user denied access")
			case errors.Is(err, oauthflowapp.ErrMissingCode):
				return httperr.BadRequest(c, "missing_code", "missing code")
			case errors.Is(err, oauthflowapp.ErrEmailNotVerified):
				return httperr.BadRequest(c, "email_not_verified", "provider did not return verified email")
			case errors.Is(err, oauthflowapp.ErrOAuthExchangeFailed):
				s.Logger.Warn("oauth exchange failed", zap.String("provider", name), zap.Error(err))
				return httperr.BadGateway(c, "oauth_exchange_failed", "oauth exchange failed")
			default:
				var link *oauthflowapp.IdentityLinkConflict
				if errors.As(err, &link) {
					return c.Status(fiber.StatusConflict).JSON(fiber.Map{
						"error": "identity_link_required", "detail": "an account exists with this email; sign in with password to link",
						"email": link.Email, "provider": link.Provider,
					})
				}
				s.Logger.Warn("oauth link failed", zap.String("provider", name), zap.Error(err))
				return httperr.Conflict(c, "oauth_link_failed", "could not link account")
			}
		}
		setHandoffCookie(c, s, res.Token)
		c.Cookie(&fiber.Cookie{Name: "eop_oauth_state", Value: "", MaxAge: -1, HTTPOnly: true, Secure: s.SecureCookies})
		dest := strings.TrimRight(s.PublicURL, "/") + "/auth/complete"
		if res.ReturnTo != "" {
			dest += "?return_to=" + base64URLEncode(res.ReturnTo)
		}
		return c.Redirect(dest, fiber.StatusFound)
	})
}

func handleSessionHandoff(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok := c.Cookies(handoffCookieName)
		if tok == "" {
			return httperr.NotFound(c, "no_handoff", "no pending session handoff")
		}

		c.Cookie(&fiber.Cookie{
			Name:     handoffCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   0,
			Expires:  time.Unix(0, 0).UTC(),
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   s.SecureCookies,
		})
		claims, err := ParseJWT(s.JWTSecret, tok)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_handoff", "invalid handoff token")
		}

		if sessionapp.HandoffAge(claims.IssuedAt, handoffTTL) {
			return httperr.Unauthorized(c, "expired_handoff", "handoff token too old")
		}
		return c.JSON(fiber.Map{"token": tok})
	}
}

func registerDevToken(g fiber.Router, s Service) {
	g.Post("/dev-token", func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		var userID uuid.UUID
		if userIDStr == "" {
			userID = uuid.New()
		} else {
			parsed, err := uuid.Parse(userIDStr)
			if err != nil {
				return httperr.BadRequest(c, "invalid_user_id", "user_id must be uuid")
			}
			userID = parsed
		}
		email := fmt.Sprintf("dev-%s@local.eop", userID.String()[:8])
		if err := s.Users.Upsert(c.Context(), userID, email, ""); err != nil {
			s.Logger.Warn("user upsert failed (continuing with token)", zap.Error(err))
		}
		tok, err := newSessionApp(s.JWTSecret, s.Pool).IssueHandoff(c.Context(), userID, email, "dev")
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"token": tok, "user_id": userID.String()})
	})
}
