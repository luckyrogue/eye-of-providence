package adminapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/audit"
	"github.com/eye-of-providence/backend/internal/teams/adminapp"
)

type fakeStore struct{}

func (fakeStore) ListTeams(context.Context, int, int) ([]adminapp.TeamRow, error) {
	return []adminapp.TeamRow{{ID: uuid.New(), Name: "t", Plan: "free", CreatedAt: time.Now()}}, nil
}

func (fakeStore) ListUsers(context.Context, int, int) ([]adminapp.UserRow, error) {
	return []adminapp.UserRow{{ID: uuid.New(), Email: "a@b.c", DisplayName: "a", GlobalRole: "user"}}, nil
}

func (fakeStore) Stats(context.Context) (adminapp.Stats, error) {
	return adminapp.Stats{UsersTotal: 1, TeamsTotal: 2, MembersTotal: 3}, nil
}

func (fakeStore) Revenue(context.Context) (adminapp.RevenueReport, error) {
	return adminapp.RevenueReport{Currency: "USD", ByPlan: map[string]int{"free": 1}}, nil
}

func (fakeStore) ListSSOConfigs(context.Context) ([]adminapp.SSOConfig, error) {
	return nil, nil
}

func (fakeStore) DisableSSO(context.Context, uuid.UUID) error {
	return adminapp.ErrSSONotConfigured
}

func (fakeStore) ListTeamPayments(context.Context, uuid.UUID) ([]adminapp.PaymentRow, error) {
	return nil, nil
}

func (fakeStore) DeleteTeam(context.Context, uuid.UUID) (adminapp.DeleteTeamResult, error) {
	return adminapp.DeleteTeamResult{TeamName: "acme"}, nil
}

func (fakeStore) DeleteUser(context.Context, uuid.UUID) (adminapp.DeleteUserResult, error) {
	return adminapp.DeleteUserResult{Email: "v@x.c", Role: "user"}, nil
}

func (fakeStore) UpdateUserRole(context.Context, uuid.UUID, string) (string, string, error) {
	return "user", "v@x.c", nil
}

func (fakeStore) UpdateUserDisplayName(context.Context, uuid.UUID, string) error {
	return nil
}

func (fakeStore) AddMember(context.Context, uuid.UUID, uuid.UUID, string) error {
	return nil
}

func (fakeStore) CountOtherOwnedTeams(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}

func (fakeStore) SetSubscription(context.Context, adminapp.SetSubscriptionInput) (adminapp.SetSubscriptionResult, error) {
	return adminapp.SetSubscriptionResult{}, nil
}

type fakeUsers struct{ id uuid.UUID }

func (f fakeUsers) FindByEmail(context.Context, string) (uuid.UUID, error) {
	return f.id, nil
}

type fakeAudit struct{}

func (fakeAudit) List(context.Context, audit.ListFilter) ([]audit.Row, error) {
	return []audit.Row{{ID: uuid.New(), Action: "test"}}, nil
}

func TestListTeams(t *testing.T) {
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}, Audit: fakeAudit{}})
	rows, err := svc.ListTeams(context.Background(), 10, 0)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

func TestRevenue(t *testing.T) {
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}})
	rep, err := svc.Revenue(context.Background())
	if err != nil || rep.Currency != "USD" {
		t.Fatalf("rep=%+v err=%v", rep, err)
	}
}

func TestDisableSSO(t *testing.T) {
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}})
	err := svc.DisableSSO(context.Background(), uuid.New())
	if !errors.Is(err, adminapp.ErrSSONotConfigured) {
		t.Fatalf("err=%v", err)
	}
}

func TestListAudit(t *testing.T) {
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}, Audit: fakeAudit{}})
	rows, err := svc.ListAudit(context.Background(), audit.ListFilter{Limit: 10})
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

func TestDeleteUserSelf(t *testing.T) {
	id := uuid.New()
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}})
	_, err := svc.DeleteUser(context.Background(), id, id)
	if !errors.Is(err, adminapp.ErrCannotDeleteSelf) {
		t.Fatalf("err=%v", err)
	}
}

func TestAddMemberInvalidEmail(t *testing.T) {
	svc := adminapp.New(adminapp.Deps{Store: fakeStore{}, Users: fakeUsers{id: uuid.New()}})
	_, err := svc.AddMember(context.Background(), adminapp.AddMemberInput{TeamID: uuid.New(), Email: "bad"})
	if !errors.Is(err, adminapp.ErrInvalidEmail) {
		t.Fatalf("err=%v", err)
	}
}
