package identitiesapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/identitiesapp"
)

type fakeRepo struct {
	rows []identitiesapp.IdentityRow
}

func (f fakeRepo) ListByUser(context.Context, uuid.UUID) ([]identitiesapp.IdentityRow, error) {
	return f.rows, nil
}

func (f fakeRepo) Delete(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

type fakeFactors struct{ n int }

func (f fakeFactors) Count(context.Context, uuid.UUID, *uuid.UUID, []byte) (int, error) {
	return f.n, nil
}

func TestList(t *testing.T) {
	id := uuid.New()
	svc := identitiesapp.New(identitiesapp.Deps{
		Repo: fakeRepo{rows: []identitiesapp.IdentityRow{{ID: id, Provider: "github", CreatedAt: time.Now()}}},
	})
	rows, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}

func TestDeleteLastFactor(t *testing.T) {
	svc := identitiesapp.New(identitiesapp.Deps{
		Repo:    fakeRepo{},
		Factors: fakeFactors{n: 0},
	})
	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, identitiesapp.ErrLastAuthFactor) {
		t.Fatalf("err=%v", err)
	}
}
