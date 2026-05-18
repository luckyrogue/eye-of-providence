package templatebc

import (
	"github.com/gofiber/fiber/v2"

	"github.com/eye-of-providence/backend/internal/_template/fooapp"
	"github.com/eye-of-providence/backend/internal/httperr"
)

// Handler — delivery layer; rename package when copying template.
type Handler struct {
	app *fooapp.Service
}

func RegisterRoutes(appRouter fiber.Router, h *Handler) {
	appRouter.Get("/example/:id", h.get)
}

func (h *Handler) get(c *fiber.Ctx) error {
	id := c.Params("id")
	ent, err := h.app.Get(c.Context(), id)
	if err != nil {
		return httperr.NotFound(c, "not_found", "not found")
	}
	return c.JSON(ent)
}
