package teamplanlimits

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/plans"
	"github.com/eye-of-providence/backend/internal/teams/teamflags"
)

type Service struct {
	store     OverrideStore
	plans     PlanDefaults
	audit     AuditSink
	limitsMap func(plans.Limits) map[string]any
}

type Deps struct {
	Store       OverrideStore
	Plans       PlanDefaults
	Audit       AuditSink
	LimitsAsMap func(plans.Limits) map[string]any
}

func New(d Deps) *Service {
	fn := d.LimitsAsMap
	if fn == nil {
		fn = defaultPlanLimitsMap
	}
	return &Service{store: d.Store, plans: d.Plans, audit: d.Audit, limitsMap: fn}
}

func defaultPlanLimitsMap(l plans.Limits) map[string]any {
	return map[string]any{
		"plan":               l.Plan,
		"max_users_per_team": l.MaxUsersPerTeam,
		"max_webhooks":       l.MaxWebhooks,
		"retention_days":     l.RetentionDays,
		"sso":                l.SSO,
		"audit_log":          l.AuditLog,
		"custom_roles":       l.CustomRoles,
		"webhook_signing":    l.WebhookSigning,
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func (s *Service) Get(ctx context.Context, teamID uuid.UUID) (*PlanLimitsView, error) {
	ov, plan, err := s.store.Read(ctx, teamID)
	if err != nil {
		return nil, err
	}
	def := s.plans.Limits(plan)
	return &PlanLimitsView{
		TeamID:            teamID,
		Plan:              plan,
		Overrides:         ov,
		EffectiveDefaults: s.limitsMap(def),
	}, nil
}

type PatchOutcome struct {
	FullReset bool
	Overrides map[string]any
}

func (s *Service) Patch(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, cmd PatchLimitsCmd) (*PatchOutcome, error) {
	if !cmd.FullReset && cmd.Patch == nil {
		return nil, ErrMissingLimits
	}
	if cmd.FullReset {
		existing, _, err := s.store.Read(ctx, teamID)
		if err != nil {
			return nil, err
		}
		if err := s.store.ClearOverride(ctx, teamID); err != nil {
			return nil, err
		}
		s.log(ctx, meta, actorID, actorEmail, "team.plan_overrides_cleared", teamID, map[string]any{
			"cleared_keys": keysOf(existing),
		})
		return &PatchOutcome{FullReset: true}, nil
	}
	normalized, verr := plans.ValidateOverrides(cmd.Patch)
	if verr != nil {
		s.logRejected(ctx, meta, actorID, actorEmail, teamID, verr)
		return nil, verr
	}
	existing, _, err := s.store.Read(ctx, teamID)
	if err != nil {
		return nil, err
	}
	merged := teamflags.PruneNullsMap(plans.MergeOverrides(existing, normalized))
	if len(merged) == 0 {
		if err := s.store.ClearOverride(ctx, teamID); err != nil {
			return nil, err
		}
	} else {
		bts, err := json.Marshal(merged)
		if err != nil {
			return nil, err
		}
		if err := s.store.SetOverrideJSON(ctx, teamID, bts); err != nil {
			return nil, err
		}
	}
	diff := plans.FlagsDiff(existing, merged)
	s.log(ctx, meta, actorID, actorEmail, "team.plan_overrides_updated", teamID, map[string]any{"diff": diff})
	return &PatchOutcome{FullReset: false, Overrides: merged}, nil
}

func (s *Service) log(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail, action string, teamID uuid.UUID, md map[string]any) {
	if s.audit == nil {
		return
	}
	s.audit.Log(ctx, AuditEvent{
		ActorID:    actorID,
		ActorEmail: actorEmail,
		Action:     action,
		TargetType: "team",
		TargetID:   teamID.String(),
		Metadata:   md,
		IP:         meta.IP,
		UserAgent:  meta.UserAgent,
	})
}

func (s *Service) logRejected(ctx context.Context, meta RequestMeta, actorID uuid.UUID, actorEmail string, teamID uuid.UUID, err error) {
	if s.audit == nil {
		return
	}
	md := map[string]any{}
	var fe *plans.FlagError
	if errors.As(err, &fe) {
		md["error_code"] = fe.Code
		md["field"] = fe.Field
	} else {
		md["error_code"] = "validation_failed"
		md["error_detail"] = err.Error()
	}
	s.log(ctx, meta, actorID, actorEmail, "team.plan_overrides_update_rejected", teamID, md)
}
