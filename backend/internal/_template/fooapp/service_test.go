package fooapp

import (
	"context"
	"testing"

	"github.com/eye-of-providence/backend/internal/_template/domain"
)

type fakeRepo struct {
	ent *domain.Entity
}

func (f fakeRepo) FindByID(context.Context, string) (*domain.Entity, error) {
	if f.ent == nil {
		return nil, domain.ErrNotFound
	}
	return f.ent, nil
}

func TestGet(t *testing.T) {
	svc := New(Deps{Repo: fakeRepo{ent: &domain.Entity{ID: "x"}}})
	ent, err := svc.Get(context.Background(), "x")
	if err != nil || ent.ID != "x" {
		t.Fatalf("got %+v err=%v", ent, err)
	}
}
