package reportsapp

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/reports/domain"
	"github.com/eye-of-providence/backend/internal/reports/periodapp"
)

const PromptVersion = "v0.1"

type Service struct {
	repo      ReportRepository
	builder   ContextBuilder
	generator ReportGenerator
	users     ActiveUsers
}

type Deps struct {
	Repo      ReportRepository
	Builder   ContextBuilder
	Generator ReportGenerator
	Users     ActiveUsers
}

func New(d Deps) *Service {
	return &Service{repo: d.Repo, builder: d.Builder, generator: d.Generator, users: d.Users}
}

func (s *Service) Generate(ctx context.Context, userID, periodKind string, now time.Time) (domain.Report, error) {
	from, to, periodKey := periodapp.Resolve(periodKind, now)
	nc, err := s.builder.Build(ctx, userID, periodKey, from, to)
	if err != nil {
		return domain.Report{}, err
	}
	body, err := s.generator.Generate(ctx, nc)
	if err != nil {
		return domain.Report{}, err
	}
	r := domain.Report{
		ID:            uuid.NewString(),
		UserID:        userID,
		Period:        periodKey,
		Model:         s.generator.Model(),
		BodyMD:        body,
		GeneratedAt:   now.UTC(),
		PromptVersion: PromptVersion,
	}
	s.repo.Save(r)
	return r, nil
}

func (s *Service) List(userID string, limit int) []domain.Report {
	if s.repo == nil {
		return nil
	}
	return s.repo.ListForUser(userID, limit)
}

func (s *Service) Get(id, userID string) (domain.Report, bool) {
	if s.repo == nil {
		return domain.Report{}, false
	}
	return s.repo.Get(id, userID)
}

func (s *Service) HasPeriod(userID, period string) bool {
	for _, r := range s.List(userID, 50) {
		if r.Period == period {
			return true
		}
	}
	return false
}

func (s *Service) ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error) {
	if s.users == nil {
		return nil, nil
	}
	return s.users.ActiveUserIDs(ctx, since)
}

func (s *Service) GenerateWeeklyIfMissing(ctx context.Context, userID string, from, to time.Time, periodKey string, now time.Time) error {
	if s.HasPeriod(userID, periodKey) {
		return nil
	}
	nc, err := s.builder.Build(ctx, userID, periodKey, from, to)
	if err != nil {
		return err
	}
	body, err := s.generator.Generate(ctx, nc)
	if err != nil {
		return err
	}
	s.repo.Save(domain.Report{
		ID:            uuid.NewString(),
		UserID:        userID,
		Period:        periodKey,
		Model:         s.generator.Model(),
		BodyMD:        body,
		GeneratedAt:   now.UTC(),
		PromptVersion: PromptVersion,
	})
	return nil
}
