package teams

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/eye-of-providence/backend/internal/httperr"
)

func adminListPagination(c *fiber.Ctx) (limit, offset int) {
	limit = 100
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, 200)
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}

func (s Service) requireSuperAdmin(c *fiber.Ctx) bool {
	if !s.isSuperAdmin(c) {
		_ = httperr.Forbidden(c, "super_admin_required", "super_admin only")
		return false
	}
	return true
}
