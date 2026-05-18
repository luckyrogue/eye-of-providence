package invitesapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
	"github.com/eye-of-providence/backend/internal/teams/invitesapp"
)

type fakeInv struct {
	team uuid.UUID
}

func (f fakeInv) FindByCode(context.Context, string) (*domain.Invite, error) {
	return &domain.Invite{TeamID: f.team, MaxUses: 2, UseCount: 0}, nil
}

func (f fakeInv) Consume(context.Context, string, uuid.UUID) (uuid.UUID, error) {
	return f.team, nil
}

func (f fakeInv) Create(context.Context, uuid.UUID, uuid.UUID, string, int, time.Time, *string) error {
	return nil
}

func (f fakeInv) MarkSent(context.Context, string, time.Time) error { return nil }

type fakeMem struct{}

func (fakeMem) AddMember(context.Context, uuid.UUID, uuid.UUID, string) error { return nil }

type fakeTeams struct {
	plan string
	n    int
}

func (f fakeTeams) Name(context.Context, uuid.UUID) (string, error) { return "Acme", nil }
func (f fakeTeams) Plan(context.Context, uuid.UUID) (string, error)  { return f.plan, nil }
func (f fakeTeams) MemberCount(context.Context, uuid.UUID) (int, error) {
	return f.n, nil
}

type fakePlans struct{}

func (fakePlans) Limits(string) invitesapp.PlanLimits {
	return invitesapp.PlanLimits{Plan: "free", MaxUsersPerTeam: 1}
}

type fakeRoles struct{ role string }

func (f fakeRoles) TeamRole(context.Context, uuid.UUID, uuid.UUID) (string, bool) {
	return f.role, f.role != ""
}

func TestAccept(t *testing.T) {
	tid := uuid.New()
	svc := invitesapp.New(invitesapp.Deps{
		Invites: fakeInv{team: tid}, Members: fakeMem{},
		Teams: fakeTeams{n: 0}, Plans: fakePlans{},
	})
	got, err := svc.Accept(context.Background(), "code", uuid.New())
	if err != nil || got != tid {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestAcceptPlanLimit(t *testing.T) {
	svc := invitesapp.New(invitesapp.Deps{
		Invites: fakeInv{team: uuid.New()}, Members: fakeMem{},
		Teams: fakeTeams{n: 5}, Plans: fakePlans{},
	})
	_, err := svc.Accept(context.Background(), "code", uuid.New())
	if !errors.Is(err, invitesapp.ErrPlanLimitExceeded) {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateRole(t *testing.T) {
	svc := invitesapp.New(invitesapp.Deps{
		Invites: fakeInv{}, Roles: fakeRoles{role: "member"},
	})
	_, err := svc.Create(context.Background(), invitesapp.CreateInput{TeamID: uuid.New(), CreatedBy: uuid.New()})
	if !errors.Is(err, invitesapp.ErrRoleInsufficient) {
		t.Fatalf("err=%v", err)
	}
}
