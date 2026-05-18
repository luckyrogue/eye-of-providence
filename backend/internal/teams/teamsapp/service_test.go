package teamsapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/teamsapp"
)

type fakeTeams struct {
	createErr error
}

func (f fakeTeams) ListForUser(context.Context, uuid.UUID) ([]teamsapp.TeamRow, error) {
	return nil, nil
}
func (f fakeTeams) GetName(context.Context, uuid.UUID) (string, error) { return "", nil }
func (f fakeTeams) Create(context.Context, teamsapp.CreateTeamParams) (uuid.UUID, error) {
	return uuid.Nil, f.createErr
}
func (f fakeTeams) UpdateName(context.Context, uuid.UUID, string) error { return nil }
func (f fakeTeams) Delete(context.Context, uuid.UUID) error             { return nil }

type fakeBeta struct{ n int }

func (f fakeBeta) TeamCount(context.Context) (int, error) { return f.n, nil }

type fakeOwners struct{ n int }

func (f fakeOwners) OwnedTeamCount(context.Context, uuid.UUID) (int, error) { return f.n, nil }

func TestCreate_OwnerLimit(t *testing.T) {
	svc := teamsapp.New(teamsapp.Deps{
		Teams: fakeTeams{}, Owners: fakeOwners{n: 1},
	})
	_, err := svc.Create(context.Background(), teamsapp.CreateInput{UserID: uuid.New(), Name: "Acme"})
	if !errors.Is(err, teamsapp.ErrOwnerLimit) {
		t.Fatalf("err=%v", err)
	}
}

func TestCreate_BetaFull(t *testing.T) {
	svc := teamsapp.New(teamsapp.Deps{
		Teams: fakeTeams{}, Beta: fakeBeta{n: 10},
	})
	_, err := svc.Create(context.Background(), teamsapp.CreateInput{
		UserID: uuid.New(), Name: "Acme", BetaLimit: 10,
	})
	if !errors.Is(err, teamsapp.ErrBetaFull) {
		t.Fatalf("err=%v", err)
	}
}
