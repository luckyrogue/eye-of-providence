package content

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/content/contentapp"
)

const (
	ActionContentPublished       = audit.ActionContentPublished
	ActionContentDraftSaved      = audit.ActionContentDraftSaved
	ActionContentRevertedDefault = audit.ActionContentRevertedDefault
	ActionContentSaveRejected    = audit.ActionContentSaveRejected
	ActionContentAccessDenied    = audit.ActionContentAccessDenied
	ActionContentPreviewAccessed = audit.ActionContentPreviewAccessed
)

const MaxContentBytes = contentapp.MaxContentBytes

// Service — delivery wiring для CMS content bounded context.
type Service struct {
	Store     *PGStore
	Cache     *Cache
	Audit     audit.Service
	Logger    *zap.Logger
	JWTSecret string

	SuperAdminCheck func(ctx context.Context, userID uuid.UUID) bool

	Pool *pgxpool.Pool

	h   *Handler
	app *contentapp.Service
}

func (s *Service) ensure() {
	if s.app != nil {
		return
	}
	s.app = newContentApp(s.Store, s.Cache, s.Audit, s.Pool, s.SuperAdminCheck)
	s.h = &Handler{app: s.app, logger: s.Logger, jwtSecret: s.JWTSecret}
}

func (s *Service) RegisterPublicRoute(app *fiber.App) {
	s.ensure()
	s.h.RegisterPublicRoute(app)
}

func (s *Service) RegisterAdminRoutes(router fiber.Router) {
	s.ensure()
	s.h.RegisterAdminRoutes(router)
}
