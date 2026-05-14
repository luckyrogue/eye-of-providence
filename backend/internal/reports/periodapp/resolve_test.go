package periodapp_test

import (
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/reports/periodapp"
)

func TestResolveWeekly(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) // Thursday
	from, to, key := periodapp.Resolve("weekly", now)
	if key == "" || !from.Before(to) {
		t.Fatalf("from=%v to=%v key=%s", from, to, key)
	}
}

func TestResolveDaily(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	_, _, key := periodapp.Resolve("daily", now)
	if key != "daily_2026-03-01" {
		t.Fatalf("got %s", key)
	}
}
