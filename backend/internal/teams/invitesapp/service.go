package invitesapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
)

type Service struct {
	invites InviteRepository
	members MemberAdder
	teams   TeamReader
	mail    InviteMailer
	plans   PlanLimiter
	roles   RoleChecker
}

type Deps struct {
	Invites InviteRepository
	Members MemberAdder
	Teams   TeamReader
	Mail    InviteMailer
	Plans   PlanLimiter
	Roles   RoleChecker
}

func New(d Deps) *Service {
	return &Service{
		invites: d.Invites,
		members: d.Members,
		teams:   d.Teams,
		mail:    d.Mail,
		plans:   d.Plans,
		roles:   d.Roles,
	}
}

type CreateInput struct {
	TeamID    uuid.UUID
	CreatedBy uuid.UUID
	Email     string
}

type CreateResult struct {
	Code      string
	ExpiresAt time.Time
	MaxUses   int
	Email     string
	Sent      bool
}

type PreviewResult struct {
	TeamID    uuid.UUID
	TeamName  string
	UsesLeft  int
	ExpiresAt *time.Time
}

func (s *Service) Create(ctx context.Context, in CreateInput) (CreateResult, error) {
	if s.invites == nil {
		return CreateResult{}, nil
	}
	if s.roles != nil {
		role, ok := s.roles.TeamRole(ctx, in.CreatedBy, in.TeamID)
		if !ok || (role != "owner" && role != "admin") {
			return CreateResult{}, ErrRoleInsufficient
		}
	}
	var emailPtr *string
	maxUses := 10
	if in.Email != "" {
		clean, ok := normalizeEmail(in.Email)
		if !ok {
			return CreateResult{}, ErrInvalidEmail
		}
		emailPtr = &clean
		maxUses = 1
	}
	code := randomCode(16)
	expires := time.Now().Add(7 * 24 * time.Hour)
	if err := s.invites.Create(ctx, in.TeamID, in.CreatedBy, code, maxUses, expires, emailPtr); err != nil {
		return CreateResult{}, err
	}
	out := CreateResult{Code: code, ExpiresAt: expires, MaxUses: maxUses}
	if emailPtr != nil {
		out.Email = *emailPtr
		if s.mail != nil {
			if err := s.mail.Send(ctx, in.TeamID, in.CreatedBy, *emailPtr, code); err == nil {
				sentAt := time.Now()
				_ = s.invites.MarkSent(ctx, code, sentAt)
				out.Sent = true
			}
		}
	}
	return out, nil
}

func (s *Service) Preview(ctx context.Context, code string) (PreviewResult, error) {
	if s.invites == nil || s.teams == nil {
		return PreviewResult{}, ErrInviteInvalid
	}
	inv, err := s.invites.FindByCode(ctx, code)
	if err != nil {
		return PreviewResult{}, ErrInviteInvalid
	}
	name, err := s.teams.Name(ctx, inv.TeamID)
	if err != nil {
		return PreviewResult{}, err
	}
	return PreviewResult{
		TeamID: inv.TeamID, TeamName: name,
		UsesLeft: inv.MaxUses - inv.UseCount, ExpiresAt: inv.Expires,
	}, nil
}

func (s *Service) Find(ctx context.Context, code string) (*domain.Invite, error) {
	if s.invites == nil {
		return nil, domain.ErrInviteInvalid
	}
	return s.invites.FindByCode(ctx, code)
}

func (s *Service) Accept(ctx context.Context, code string, userID uuid.UUID) (uuid.UUID, error) {
	if s.invites == nil || s.members == nil {
		return uuid.Nil, domain.ErrInviteInvalid
	}
	inv, err := s.invites.FindByCode(ctx, code)
	if err != nil {
		return uuid.Nil, ErrInviteInvalid
	}
	if err := s.checkTeamCapacity(ctx, inv.TeamID); err != nil {
		return uuid.Nil, err
	}
	teamID, err := s.invites.Consume(ctx, code, userID)
	if err != nil {
		return uuid.Nil, ErrInviteInvalid
	}
	if err := s.members.AddMember(ctx, teamID, userID, domain.RoleMember); err != nil {
		return uuid.Nil, err
	}
	return teamID, nil
}

func (s *Service) checkTeamCapacity(ctx context.Context, teamID uuid.UUID) error {
	if s.teams == nil || s.plans == nil {
		return nil
	}
	plan, err := s.teams.Plan(ctx, teamID)
	if err != nil {
		return err
	}
	limits := s.plans.Limits(plan)
	if limits.MaxUsersPerTeam <= 0 {
		return nil
	}
	count, err := s.teams.MemberCount(ctx, teamID)
	if err != nil {
		return err
	}
	if count >= limits.MaxUsersPerTeam {
		return fmt.Errorf("%w: team is on %s plan, limited to %d members (current: %d)",
			ErrPlanLimitExceeded, limits.Plan, limits.MaxUsersPerTeam, count)
	}
	return nil
}

func randomCode(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
