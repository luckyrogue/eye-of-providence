package teamflags

import (
	"context"

	"github.com/google/uuid"
)

type FlagStore interface {
	Load(ctx context.Context, teamID uuid.UUID) (map[string]any, error)
	Save(ctx context.Context, teamID uuid.UUID, flagsJSON []byte) (rowsAffected int64, err error)
}

type AuditSink interface {
	Log(ctx context.Context, e AuditEvent)
}

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

type RequestMeta struct {
	IP        string
	UserAgent string
}
