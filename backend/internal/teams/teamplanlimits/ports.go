package teamplanlimits

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/plans"
)

type OverrideStore interface {
	Read(ctx context.Context, teamID uuid.UUID) (override map[string]any, plan string, err error)
	ClearOverride(ctx context.Context, teamID uuid.UUID) error
	SetOverrideJSON(ctx context.Context, teamID uuid.UUID, json []byte) error
}

type PlanDefaults interface {
	Limits(plan string) plans.Limits
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

type PlanLimitsView struct {
	TeamID            uuid.UUID
	Plan              string
	Overrides         map[string]any
	EffectiveDefaults map[string]any
}

type PatchLimitsCmd struct {
	FullReset bool
	Patch     map[string]any
}
