package teamsapp

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

const MaxTeamNameLen = 100

type Service struct {
	teams  TeamRepository
	beta   BetaGate
	owners OwnerLimitChecker
}

type Deps struct {
	Teams  TeamRepository
	Beta   BetaGate
	Owners OwnerLimitChecker
}

func New(d Deps) *Service {
	return &Service{teams: d.Teams, beta: d.Beta, owners: d.Owners}
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID) ([]TeamRow, error) {
	if s.teams == nil {
		return []TeamRow{}, nil
	}
	return s.teams.ListForUser(ctx, userID)
}

type CreateInput struct {
	UserID    uuid.UUID
	Name      string
	IsSuper   bool
	BetaLimit int
	LockID    int64
}

type CreateResult struct {
	ID   uuid.UUID
	Name string
	Role string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (CreateResult, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > MaxTeamNameLen {
		return CreateResult{}, ErrInvalidName
	}
	if !in.IsSuper && s.owners != nil {
		n, err := s.owners.OwnedTeamCount(ctx, in.UserID)
		if err != nil {
			return CreateResult{}, err
		}
		if n > 0 {
			return CreateResult{}, ErrOwnerLimit
		}
	}
	if in.BetaLimit > 0 && !in.IsSuper && s.beta != nil {
		cnt, err := s.beta.TeamCount(ctx)
		if err != nil {
			return CreateResult{}, err
		}
		if cnt >= in.BetaLimit {
			return CreateResult{}, ErrBetaFull
		}
	}
	if s.teams == nil {
		return CreateResult{}, nil
	}
	id, err := s.teams.Create(ctx, CreateTeamParams{
		UserID: in.UserID, Name: name, IsSuper: in.IsSuper, BetaLimit: in.BetaLimit, LockID: in.LockID,
	})
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{ID: id, Name: name, Role: "owner"}, nil
}

type DetailResult struct {
	ID   uuid.UUID
	Name string
	Role string
}

func (s *Service) GetDetail(ctx context.Context, userID, teamID uuid.UUID, role string) (DetailResult, error) {
	if s.teams == nil {
		return DetailResult{ID: teamID, Role: role}, nil
	}
	name, err := s.teams.GetName(ctx, teamID)
	if err != nil {
		return DetailResult{}, ErrTeamNotFound
	}
	return DetailResult{ID: teamID, Name: name, Role: role}, nil
}

func (s *Service) Update(ctx context.Context, teamID uuid.UUID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > MaxTeamNameLen {
		return ErrInvalidName
	}
	if s.teams == nil {
		return nil
	}
	return s.teams.UpdateName(ctx, teamID, name)
}

func (s *Service) Delete(ctx context.Context, teamID uuid.UUID) error {
	if s.teams == nil {
		return nil
	}
	return s.teams.Delete(ctx, teamID)
}

type BetaInfoResult struct {
	TeamsCount     int
	Limit          int
	SlotsRemaining int
}

func (s *Service) BetaInfo(ctx context.Context, limit int) (BetaInfoResult, error) {
	out := BetaInfoResult{Limit: limit, SlotsRemaining: -1}
	if s.beta == nil {
		return out, nil
	}
	cnt, err := s.beta.TeamCount(ctx)
	if err != nil {
		return BetaInfoResult{}, err
	}
	out.TeamsCount = cnt
	if limit > 0 {
		rem := limit - cnt
		if rem < 0 {
			rem = 0
		}
		out.SlotsRemaining = rem
	}
	return out, nil
}
