package batchapp_test

import (
	"testing"

	"github.com/eye-of-providence/backend/internal/ingest/batchapp"
	"github.com/eye-of-providence/backend/internal/store"
)

func TestPrepareIngest_TooLarge(t *testing.T) {
	ev := make([]store.Event, 3)
	_, _, _, err := batchapp.PrepareIngest("u", ev, 2)
	if err != batchapp.ErrBatchTooLarge {
		t.Fatalf("got %v", err)
	}
}

func TestPrepareIngest_AcceptReject(t *testing.T) {
	ev := []store.Event{
		{AppBundle: "b", Source: "ide", Category: "ai", DurationMS: 1},
		{AppBundle: "", Source: "ide", Category: "ai"},
	}
	out, acc, rej, err := batchapp.PrepareIngest("user-1", ev, 10)
	if err != nil || acc != 1 || rej != 1 || len(out) != 1 || out[0].UserID != "user-1" {
		t.Fatalf("out=%v acc=%d rej=%d err=%v", out, acc, rej, err)
	}
}

func TestValidEvent_Duration(t *testing.T) {
	e := store.Event{AppBundle: "b", Source: "cli", Category: "other", DurationMS: 25 * 60 * 60 * 1000}
	if batchapp.ValidEvent(e) {
		t.Fatal("expected reject")
	}
}
