package adminapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
)

type Service struct {
	store       Store
	audit       AuditLister
	userDeleter UserDeleter
	tokenBumper TokenBumper
	users       UserFinder
}

type Deps struct {
	Store       Store
	Audit       AuditLister
	UserDeleter UserDeleter
	TokenBumper TokenBumper
	Users       UserFinder
}

func New(d Deps) *Service {
	return &Service{
		store:       d.Store,
		audit:       d.Audit,
		userDeleter: d.UserDeleter,
		tokenBumper: d.TokenBumper,
		users:       d.Users,
	}
}

func (s *Service) ListTeams(ctx context.Context, limit, offset int) ([]TeamRow, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.ListTeams(ctx, limit, offset)
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]UserRow, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.ListUsers(ctx, limit, offset)
}

func (s *Service) Stats(ctx context.Context) (Stats, error) {
	if s.store == nil {
		return Stats{}, nil
	}
	return s.store.Stats(ctx)
}

func (s *Service) Revenue(ctx context.Context) (RevenueReport, error) {
	if s.store == nil {
		return RevenueReport{Currency: "USD", ByPlan: map[string]int{}}, nil
	}
	return s.store.Revenue(ctx)
}

func (s *Service) ListSSOConfigs(ctx context.Context) ([]SSOConfig, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.ListSSOConfigs(ctx)
}

func (s *Service) DisableSSO(ctx context.Context, teamID uuid.UUID) error {
	if s.store == nil {
		return ErrSSONotConfigured
	}
	return s.store.DisableSSO(ctx, teamID)
}

func (s *Service) ListTeamPayments(ctx context.Context, teamID uuid.UUID) ([]PaymentRow, error) {
	if s.store == nil {
		return nil, nil
	}
	return s.store.ListTeamPayments(ctx, teamID)
}

func (s *Service) DeleteTeam(ctx context.Context, teamID uuid.UUID) (DeleteTeamResult, error) {
	if s.store == nil {
		return DeleteTeamResult{}, nil
	}
	return s.store.DeleteTeam(ctx, teamID)
}

func (s *Service) ListAudit(ctx context.Context, f audit.ListFilter) ([]audit.Row, error) {
	if s.audit == nil {
		return nil, nil
	}
	return s.audit.List(ctx, f)
}
