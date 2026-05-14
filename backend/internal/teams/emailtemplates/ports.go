package emailtemplates

import (
	"context"

	"github.com/google/uuid"
)

// OverrideRepository — Postgres overrides для transactional templates.
type OverrideRepository interface {
	ListOverrides(ctx context.Context) ([]OverrideRow, error)
	Lookup(ctx context.Context, key, locale string) (*OverrideRow, error)
	Upsert(ctx context.Context, row OverrideRow, actorID uuid.UUID) (*OverrideRow, error)
	Delete(ctx context.Context, key, locale string) (*OverrideRow, error)
}

// BaselineProvider — embedded baseline когда нет DB row.
type BaselineProvider interface {
	Template(key, locale string) *OverrideRow
}

// TemplateSyntaxValidator — sanity-check шаблонов при save (пустой vars).
type TemplateSyntaxValidator interface {
	Validate(subject, bodyHTML, bodyText string, key, locale string) error
}

// AuditSink — append-only audit без зависимости от Fiber.
type AuditSink interface {
	Log(ctx context.Context, e AuditEvent)
}
