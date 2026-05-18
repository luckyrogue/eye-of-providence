package membersapp_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
	"github.com/eye-of-providence/backend/internal/teams/membersapp"
)

type fakeRepo struct{}

func (fakeRepo) ListByTeam(context.Context, domain.TeamID) ([]domain.Member, error) {
	return []domain.Member{{ID: uuid.New(), Email: "a@b.c", DisplayName: "A", Role: domain.RoleMember, JoinedAt: time.Now()}}, nil
}

func TestList(t *testing.T) {
	svc := membersapp.New(membersapp.Deps{Members: fakeRepo{}})
	out, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
