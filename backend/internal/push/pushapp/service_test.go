package pushapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/domain"
	"github.com/eye-of-providence/backend/internal/push/pushapp"
)

type fakeRepo struct{}

func (fakeRepo) List(context.Context, uuid.UUID) ([]domain.Subscription, error) {
	return []domain.Subscription{{Endpoint: "https://push"}}, nil
}

func (fakeRepo) Subscribe(context.Context, uuid.UUID, string, string, string, string) error {
	return nil
}

func (fakeRepo) Unsubscribe(context.Context, uuid.UUID, string) (bool, error) {
	return true, nil
}

func TestList(t *testing.T) {
	svc := pushapp.New(pushapp.Deps{Repo: fakeRepo{}})
	out, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
