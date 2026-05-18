package membersapp

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type Service struct {
	members  MemberRepository
	roles    RoleReader
	mutator  MemberMutator
	activity ActivityAggregator
}

type Deps struct {
	Members  MemberRepository
	Roles    RoleReader
	Mutator  MemberMutator
	Activity ActivityAggregator
}

func New(d Deps) *Service {
	return &Service{
		members:  d.Members,
		roles:    d.Roles,
		mutator:  d.Mutator,
		activity: d.Activity,
	}
}

func (s *Service) List(ctx context.Context, teamID domain.TeamID) ([]domain.Member, error) {
	if s.members == nil {
		return []domain.Member{}, nil
	}
	return s.members.ListByTeam(ctx, teamID)
}

func (s *Service) TeamRole(ctx context.Context, userID, teamID uuid.UUID) (string, bool) {
	if s.roles == nil {
		return "", false
	}
	return s.roles.TeamRole(ctx, userID, teamID)
}

type SummaryMember struct {
	ID          uuid.UUID
	DisplayName string
	AIMS        uint64
	ManualMS    uint64
	TotalMS     uint64
	AIRatio     int
}

type TeamSummaryResult struct {
	Members []SummaryMember
	Since   time.Time
}

func (s *Service) TeamSummary(ctx context.Context, teamID uuid.UUID) (TeamSummaryResult, error) {
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	if s.members == nil {
		return TeamSummaryResult{Since: since}, nil
	}
	members, err := s.members.ListByTeam(ctx, teamID)
	if err != nil {
		return TeamSummaryResult{}, err
	}
	out := make([]SummaryMember, 0, len(members))
	ids := make([]string, 0, len(members))
	for _, m := range members {
		out = append(out, SummaryMember{ID: m.ID, DisplayName: m.DisplayName})
		ids = append(ids, m.ID.String())
	}
	if s.activity != nil && len(ids) > 0 {
		bulk, err := s.activity.AggregateByCategoryBulk(ctx, ids, since)
		if err == nil {
			for i := range out {
				agg := bulk[out[i].ID.String()]
				out[i].AIMS = agg["ai"]
				out[i].ManualMS = agg["manual"] + agg["refactor"]
				out[i].TotalMS = out[i].AIMS + out[i].ManualMS + agg["other"] + agg["reading"]
				if out[i].TotalMS > 0 {
					out[i].AIRatio = int(float64(out[i].AIMS) * 100.0 / float64(out[i].TotalMS))
				}
			}
		}
	}
	return TeamSummaryResult{Members: out, Since: since}, nil
}

type UpdateRoleInput struct {
	ActorID      uuid.UUID
	TeamID       uuid.UUID
	TargetUserID uuid.UUID
	NewRole      string
	AllowOwner   bool // super_admin bypass for owner limit
}

func (s *Service) UpdateRole(ctx context.Context, in UpdateRoleInput) error {
	if s.roles == nil || s.mutator == nil {
		return nil
	}
	role, ok := s.roles.TeamRole(ctx, in.ActorID, in.TeamID)
	if !ok || role != "owner" {
		return ErrOwnerRequired
	}
	newRole := strings.ToLower(strings.TrimSpace(in.NewRole))
	if newRole != "owner" && newRole != "admin" && newRole != "member" {
		return ErrInvalidRole
	}
	if newRole == "owner" && !in.AllowOwner {
		n, err := s.mutator.OwnedTeamCount(ctx, in.TargetUserID, in.TeamID)
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrOwnerLimit
		}
	}
	if in.ActorID == in.TargetUserID && newRole != "owner" {
		cnt, err := s.mutator.OwnerCount(ctx, in.TeamID)
		if err != nil {
			return err
		}
		if cnt <= 1 {
			return ErrLastOwner
		}
	}
	return s.mutator.UpdateRole(ctx, in.TeamID, in.TargetUserID, newRole)
}

type RemoveMemberInput struct {
	ActorID      uuid.UUID
	TeamID       uuid.UUID
	TargetUserID uuid.UUID
}

func (s *Service) Remove(ctx context.Context, in RemoveMemberInput) error {
	if s.roles == nil || s.mutator == nil {
		return nil
	}
	role, ok := s.roles.TeamRole(ctx, in.ActorID, in.TeamID)
	if !ok || (role != "owner" && role != "admin") {
		return ErrRoleInsufficient
	}
	targetRole, err := s.mutator.TargetRole(ctx, in.TeamID, in.TargetUserID)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		cnt, err := s.mutator.OwnerCount(ctx, in.TeamID)
		if err != nil {
			return err
		}
		if cnt <= 1 {
			return ErrLastOwner
		}
	}
	if role == "admin" && targetRole == "owner" {
		return ErrAdminCantRemove
	}
	return s.mutator.Remove(ctx, in.TeamID, in.TargetUserID)
}
