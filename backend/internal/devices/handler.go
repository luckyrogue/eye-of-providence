package devices

import (
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
	Pool      *pgxpool.Pool
	Logger    *zap.Logger
	JWTSecret string
}

func RegisterRoutes(app *fiber.App, s Service) {
	if s.Pool == nil {

		app.Post("/v1/devices/pair", func(c *fiber.Ctx) error {
			return httperr.Unavailable(c, "db_not_configured", "device pairing requires database")
		})
		app.Post("/v1/devices/poll", func(c *fiber.Ctx) error {
			return httperr.Unavailable(c, "db_not_configured", "device pairing requires database")
		})
		return
	}

	app.Post("/v1/devices/pair", pairBeginHandler(s))
	app.Post("/v1/devices/poll", pollHandler(s))

	g := app.Group("/v1/me/devices", auth.Middleware(s.JWTSecret, s.Pool))
	g.Get("/", listHandler(s))
	g.Post("/claim", claimHandler(s))
	g.Delete("/:id", revokeHandler(s))
}

func pairBeginHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		req.Kind = strings.ToLower(strings.TrimSpace(req.Kind))
		res, err := newPairingApp(s.Pool).Begin(c.Context(), req.Kind, req.Name)
		if errors.Is(err, ErrInvalidKind) {
			return httperr.BadRequest(c, "invalid_kind", "kind must be one of ext|agent|ide")
		}
		if err != nil {
			s.Logger.Error("pair begin failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(res)
	}
}

func pollHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			PairID string `json:"pair_id"`
			Secret string `json:"secret"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		pid, err := uuid.Parse(strings.TrimSpace(req.PairID))
		if err != nil {
			return httperr.BadRequest(c, "invalid_pair_id", "pair_id must be uuid")
		}
		res, err := newPairingApp(s.Pool).Poll(c.Context(), pid, req.Secret)
		if errors.Is(err, ErrPairingNotFound) {
			return httperr.NotFound(c, "pairing_not_found", "pairing not found or expired")
		}
		if errors.Is(err, ErrSecretMismatch) {
			return httperr.Forbidden(c, "secret_mismatch", "pairing secret mismatch")
		}
		if err != nil {
			s.Logger.Error("pair poll failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(res)
	}
}

func claimHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if auth.ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "device claim requires JWT (dashboard session)")
		}
		claims := auth.ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		var req struct {
			Code string `json:"code"`
			Name string `json:"name"`
		}
		if err := c.BodyParser(&req); err != nil {
			return httperr.BadRequest(c, "invalid_body", "invalid body")
		}
		dev, err := newPairingApp(s.Pool).Claim(c.Context(), uid, req.Code, req.Name)
		if errors.Is(err, ErrCodeNotFound) {
			return httperr.NotFound(c, "code_not_found", "code invalid or expired")
		}
		if errors.Is(err, ErrAlreadyClaimed) {
			return httperr.Conflict(c, "code_already_claimed", "code already used")
		}
		if err != nil {
			s.Logger.Error("device claim failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(dev)
	}
}

func listHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if auth.ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "devices list requires JWT")
		}
		claims := auth.ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		devs, err := newDevicesApp(s.Pool).List(c.Context(), uid)
		if err != nil {
			s.Logger.Error("devices list failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{"devices": devs})
	}
}

func revokeHandler(s Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if auth.ScopeFromCtx(c) != "" {
			return httperr.Forbidden(c, "jwt_required", "device revoke requires JWT")
		}
		claims := auth.ClaimsFromCtx(c)
		uid, err := uuid.Parse(claims.UserID)
		if err != nil {
			return httperr.Unauthorized(c, "invalid_subject", "invalid subject")
		}
		deviceID, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return httperr.BadRequest(c, "invalid_id", "invalid device id")
		}
		ok, err := newDevicesApp(s.Pool).Revoke(c.Context(), uid, deviceID)
		if err != nil {
			s.Logger.Error("device revoke failed", zap.Error(err))
			return httperr.Internal(c)
		}
		if !ok {
			return httperr.NotFound(c, "device_not_found", "device not found")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	}
}
