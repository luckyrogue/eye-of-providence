package auth

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
)

func handleWebAuthnRegisterBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		wa := newWebAuthnApp(s.WebAuthn)
		creation, sid, err := wa.RegisterBegin(c.Context(), uid)
		if err != nil {
			s.Logger.Warn("webauthn begin register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    creation,
			"session_id": sid,
		})
	}
}

type webauthnFinishRegisterReq struct {
	SessionID   string          `json:"session_id"`
	Attestation json.RawMessage `json:"attestation"`
	Nickname    string          `json:"nickname"`
}

func handleWebAuthnRegisterFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid token subject")
		}
		var req webauthnFinishRegisterReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Attestation) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and attestation required")
		}
		if err := newWebAuthnApp(s.WebAuthn).RegisterFinish(c.Context(), uid, req.SessionID, req.Attestation, req.Nickname); err != nil {
			s.Logger.Warn("webauthn finish register failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_finish_failed", err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

type webauthnLoginBeginReq struct {
	Email *string `json:"email"`
}

func handleWebAuthnLoginBegin(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginBeginReq
		_ = c.BodyParser(&req)
		assertion, sid, err := newWebAuthnApp(s.WebAuthn).LoginBegin(c.Context(), req.Email)
		if err != nil {
			s.Logger.Warn("webauthn begin login failed", zap.Error(err))
			return httperr.BadRequest(c, "webauthn_begin_failed", err.Error())
		}
		return c.JSON(fiber.Map{
			"options":    assertion,
			"session_id": sid,
		})
	}
}

type webauthnLoginFinishReq struct {
	SessionID string          `json:"session_id"`
	Assertion json.RawMessage `json:"assertion"`
	ReturnTo  string          `json:"return_to"`
}

func handleWebAuthnLoginFinish(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req webauthnLoginFinishReq
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		if req.SessionID == "" || len(req.Assertion) == 0 {
			return httperr.BadRequest(c, "missing_fields", "session_id and assertion required")
		}
		uid, err := newWebAuthnApp(s.WebAuthn).LoginFinish(c.Context(), req.SessionID, req.Assertion)
		if err != nil {
			s.Logger.Warn("webauthn finish login failed", zap.Error(err))
			return httperr.Unauthorized(c, "webauthn_finish_failed", err.Error())
		}

		tok, err := newSessionApp(s.JWTSecret, s.Pool).IssueHandoff(c.Context(), uid, "", "passkey")
		if err != nil {
			s.Logger.Error("issue jwt failed", zap.Error(err))
			return httperr.Internal(c)
		}
		setHandoffCookie(c, s, tok)
		return c.JSON(fiber.Map{"token": tok, "user_id": uid.String()})
	}
}
