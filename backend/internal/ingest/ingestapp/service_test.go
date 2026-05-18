package ingestapp_test

import (
	"testing"

	"github.com/eye-of-providence/backend/internal/ingest/domain"
	"github.com/eye-of-providence/backend/internal/ingest/ingestapp"
)

func TestPrepareBatch_TooLarge(t *testing.T) {
	ev := make([]domain.Event, 3)
	_, _, err := ingestapp.New(ingestapp.Deps{}).PrepareBatch("u", ev, 2)
	if err != domain.ErrBatchTooLarge {
		t.Fatalf("got %v", err)
	}
}

func TestPrepareBatch_AcceptReject(t *testing.T) {
	ev := []domain.Event{
		{AppBundle: "b", Source: "ide", Category: "ai", DurationMS: 1},
		{AppBundle: "", Source: "ide", Category: "ai"},
	}
	out, res, err := ingestapp.New(ingestapp.Deps{}).PrepareBatch("user-1", ev, 10)
	if err != nil || res.Accepted != 1 || res.Rejected != 1 || len(out) != 1 || out[0].UserID != "user-1" {
		t.Fatalf("out=%v res=%+v err=%v", out, res, err)
	}
}

func TestValidEvent_Duration(t *testing.T) {
	e := domain.Event{AppBundle: "b", Source: "cli", Category: "other", DurationMS: 25 * 60 * 60 * 1000}
	if domain.ValidEvent(e) {
		t.Fatal("expected reject")
	}
}
