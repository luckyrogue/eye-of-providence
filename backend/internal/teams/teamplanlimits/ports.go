package teamplanlimits

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/plans"
)

// OverrideStore — read/write plan_limits_override + subscription_plan.
type OverrideStore interface {
	Read(ctx context.Context, teamID uuid.UUID) (override map[string]any, plan string, err error)
	ClearOverride(ctx context.Context, teamID uuid.UUID) error
	SetOverrideJSON(ctx context.Context, teamID uuid.UUID, json []byte) error
}

// PlanDefaults — effective limits для named plan.
type PlanDefaults interface {
	Limits(plan string) plans.Limits
}

// AuditSink — append-only audit.
type AuditSink interface {
	Log(ctx context.Context, e AuditEvent)
}

// AuditEvent — audit row.
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

// RequestMeta — IP/UA.
type RequestMeta struct {
	IP        string
	UserAgent string
}

// PlanLimitsView — ответ GET.
type PlanLimitsView struct {
	TeamID            uuid.UUID
	Plan              string
	Overrides         map[string]any
	EffectiveDefaults map[string]any
}

// PatchLimitsCmd — результат разбора PATCH body (handler парсит JSON).
type PatchLimitsCmd struct {
	FullReset bool           // `"limits": null`
	Patch     map[string]any // non-nil object
}
