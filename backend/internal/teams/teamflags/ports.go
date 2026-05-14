package teamflags

import (
	"context"

	"github.com/google/uuid"
)

// FlagStore — load/save teams.flags JSONB.
type FlagStore interface {
	Load(ctx context.Context, teamID uuid.UUID) (map[string]any, error)
	Save(ctx context.Context, teamID uuid.UUID, flagsJSON []byte) (rowsAffected int64, err error)
}

// AuditSink — append-only audit без Fiber.
type AuditSink interface {
	Log(ctx context.Context, e AuditEvent)
}

// AuditEvent — запись audit trail.
type AuditEvent struct {
	ActorID    uuid.UUID
	ActorEmail string
	Action     string
	TargetType string
	TargetID   string
	Metadata   map[string]any
	IP         string
	UserAgent  string
}

// RequestMeta — IP/UA для audit.
type RequestMeta struct {
	IP        string
	UserAgent string
}
