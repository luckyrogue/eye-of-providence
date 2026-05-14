package teams

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/plans"
	"github.com/eye-of-providence/backend/internal/teams/teamplanlimits"
)

type pgPlanOverrideStore struct {
	pool *pgxpool.Pool
}

func (s pgPlanOverrideStore) Read(ctx context.Context, teamID uuid.UUID) (map[string]any, string, error) {
	if s.pool == nil {
		return map[string]any{}, "free", nil
	}
	var raw []byte
	var plan string
	err := s.pool.QueryRow(ctx,
		"SELECT plan_limits_override, COALESCE(subscription_plan, 'free') FROM teams WHERE id = $1",
		teamID).Scan(&raw, &plan)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", teamplanlimits.ErrTeamNotFound
	}
	if err != nil {
		return nil, "", err
	}
	out := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, plan, nil
}

func (s pgPlanOverrideStore) ClearOverride(ctx context.Context, teamID uuid.UUID) error {
	if s.pool == nil {
		return errors.New("nil pool")
	}
	_, err := s.pool.Exec(ctx, "UPDATE teams SET plan_limits_override = NULL WHERE id = $1", teamID)
	return err
}

func (s pgPlanOverrideStore) SetOverrideJSON(ctx context.Context, teamID uuid.UUID, json []byte) error {
	if s.pool == nil {
		return errors.New("nil pool")
	}
	_, err := s.pool.Exec(ctx, "UPDATE teams SET plan_limits_override = $1 WHERE id = $2", json, teamID)
	return err
}

type planLimitsDefaultsAdapter struct {
	svc plans.Service
}

func (a planLimitsDefaultsAdapter) Limits(plan string) plans.Limits {
	return a.svc.Limits(plan)
}

type teamplanlimitsAuditAdapter struct {
	svc audit.Service
}

func (a teamplanlimitsAuditAdapter) Log(ctx context.Context, e teamplanlimits.AuditEvent) {
	if a.svc.Pool == nil {
		return
	}
	a.svc.Log(ctx, audit.Entry{
		ActorID:    e.ActorID,
		ActorEmail: e.ActorEmail,
		Action:     audit.Action(e.Action),
		TargetType: e.TargetType,
		TargetID:   e.TargetID,
		Metadata:   e.Metadata,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
	})
}

func (s Service) newTeamPlanLimitsService() *teamplanlimits.Service {
	return teamplanlimits.New(teamplanlimits.Deps{
		Store:       pgPlanOverrideStore{pool: s.Pool},
		Plans:       planLimitsDefaultsAdapter{svc: s.Plans},
		Audit:       teamplanlimitsAuditAdapter{svc: s.Audit},
		LimitsAsMap: nil, // use package default
	})
}
