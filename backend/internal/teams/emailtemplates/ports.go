package emailtemplates

import (
	"context"

	"github.com/google/uuid"
)

type OverrideRepository interface {
	ListOverrides(ctx context.Context) ([]OverrideRow, error)
	Lookup(ctx context.Context, key, locale string) (*OverrideRow, error)
	Upsert(ctx context.Context, row OverrideRow, actorID uuid.UUID) (*OverrideRow, error)
	Delete(ctx context.Context, key, locale string) (*OverrideRow, error)
}

type BaselineProvider interface {
	Template(key, locale string) *OverrideRow
}

type TemplateSyntaxValidator interface {
	Validate(subject, bodyHTML, bodyText string, key, locale string) error
}

type AuditSink interface {
	Log(ctx context.Context, e AuditEvent)
}
