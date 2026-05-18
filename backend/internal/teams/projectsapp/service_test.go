package projectsapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/teams/domain"
	"github.com/eye-of-providence/backend/internal/teams/projectsapp"
)

type fakeProj struct{}

func (fakeProj) List(context.Context, uuid.UUID) ([]domain.Project, error) {
	return []domain.Project{{Name: "p1"}}, nil
}

func (fakeProj) Create(context.Context, uuid.UUID, uuid.UUID, domain.CreateProjectInput) (uuid.UUID, error) {
	return uuid.New(), nil
}

func TestList(t *testing.T) {
	out, err := projectsapp.New(projectsapp.Deps{Repo: fakeProj{}}).List(context.Background(), uuid.New())
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
