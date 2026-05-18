package devicesapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/devicesapp"
	"github.com/eye-of-providence/backend/internal/devices/domain"
)

type fakeRepo struct{}

func (fakeRepo) List(context.Context, uuid.UUID) ([]domain.Device, error) {
	return []domain.Device{{Name: "laptop"}}, nil
}

func (fakeRepo) Revoke(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func TestList(t *testing.T) {
	svc := devicesapp.New(devicesapp.Deps{Repo: fakeRepo{}})
	out, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
