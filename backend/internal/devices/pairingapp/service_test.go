package pairingapp_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/devices/domain"
	"github.com/eye-of-providence/backend/internal/devices/pairingapp"
)

type fakePair struct{}

func (fakePair) Begin(context.Context, string, string) (domain.PairBeginResult, error) {
	return domain.PairBeginResult{Code: "ABC123"}, nil
}

func (fakePair) Poll(context.Context, uuid.UUID, string) (domain.PollResult, error) {
	return domain.PollResult{Status: "pending"}, nil
}

func (fakePair) Claim(context.Context, uuid.UUID, string, string) (domain.Device, error) {
	return domain.Device{Name: "dev"}, nil
}

func TestBegin(t *testing.T) {
	svc := pairingapp.New(pairingapp.Deps{Repo: fakePair{}})
	res, err := svc.Begin(context.Background(), "agent", "")
	if err != nil || res.Code != "ABC123" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}
