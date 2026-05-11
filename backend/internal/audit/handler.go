package audit

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/httperr"
)

// RegisterRoutes — admin audit endpoints под /v1/admin/audit. Super_admin
// only — проверка через requireSuper callback (caller передаёт
// teams.Service.isSuperAdmin или эквивалент).
//
// Endpoints:
//
//	GET /v1/admin/audit?action=...&target_type=...&target_id=...&actor_id=...&limit=N&offset=N
//	  → {entries: [...], limit, offset}
func RegisterRoutes(app *fiber.App, jwtSecret string, s Service, requireSuper func(c *fiber.Ctx) bool) {
	app.Get("/v1/admin/audit", func(c *fiber.Ctx) error {
		if !requireSuper(c) {
			return httperr.Forbidden(c, "super_admin_required", "super_admin role required")
		}
		f := ListFilter{
			Action:     c.Query("action"),
			TargetType: c.Query("target_type"),
			TargetID:   c.Query("target_id"),
		}
		if a := c.Query("actor_id"); a != "" {
			id, err := uuid.Parse(a)
			if err != nil {
				return httperr.BadRequest(c, "invalid_actor_id", "actor_id must be uuid")
			}
			f.ActorID = id
		}
		if l := c.Query("limit"); l != "" {
			n, err := strconv.Atoi(l)
			if err == nil && n > 0 {
				f.Limit = n
			}
		}
		if o := c.Query("offset"); o != "" {
			n, err := strconv.Atoi(o)
			if err == nil && n > 0 {
				f.Offset = n
			}
		}
		entries, err := s.List(c.Context(), f)
		if err != nil {
			s.Logger.Error("audit list failed", zap.Error(err))
			return httperr.Internal(c)
		}
		return c.JSON(fiber.Map{
			"entries": entries,
			"limit":   f.Limit,
			"offset":  f.Offset,
		})
	})
	// jwtSecret currently unused but сохраняем сигнатуру для consistency
	// с другими RegisterRoutes — guard выполняется через requireSuper,
	// auth.Middleware уже навешан caller'ом через main.go.
	_ = jwtSecret
}
