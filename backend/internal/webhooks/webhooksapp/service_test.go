package webhooksapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/domain"
	"github.com/eye-of-providence/backend/internal/webhooks/webhooksapp"
)

type fakeRepo struct{}

func (fakeRepo) List(context.Context, uuid.UUID) ([]domain.Webhook, error) {
	return []domain.Webhook{{URL: "https://x"}}, nil
}

func (fakeRepo) Create(context.Context, uuid.UUID, string, []string, string) (string, domain.Webhook, error) {
	return "sec", domain.Webhook{URL: "https://x"}, nil
}

func (fakeRepo) Delete(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func TestList(t *testing.T) {
	svc := webhooksapp.New(webhooksapp.Deps{Repo: fakeRepo{}})
	out, err := svc.List(context.Background(), uuid.New())
	if err != nil || len(out) != 1 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
