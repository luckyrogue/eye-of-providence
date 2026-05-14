package batchapp

import (
	"errors"
	"time"

	"github.com/eye-of-providence/backend/internal/store"
)

// ErrBatchTooLarge — превышен лимит событий в одном запросе.
var ErrBatchTooLarge = errors.New("batch too large")

// PrepareIngest — валидация размера batch + privacy-gate + user_id из токена.
func PrepareIngest(claimsUserID string, events []store.Event, maxBatch int) ([]store.Event, int, int, error) {
	if len(events) > maxBatch {
		return nil, 0, 0, ErrBatchTooLarge
	}
	now := time.Now().UTC()
	valid := make([]store.Event, 0, len(events))
	accepted, rejected := 0, 0
	for _, e := range events {
		if !ValidEvent(e) {
			rejected++
			continue
		}
		e.UserID = claimsUserID
		if e.TS.IsZero() {
			e.TS = now
		}
		valid = append(valid, e)
		accepted++
	}
	return valid, accepted, rejected, nil
}

// ValidEvent — privacy-gate и базовые инварианты (см. ingest handler).
func ValidEvent(e store.Event) bool {
	if e.AppBundle == "" || e.Source == "" || e.Category == "" {
		return false
	}
	switch e.Source {
	case "os", "browser", "ide", "cli":
	default:
		return false
	}
	switch e.Category {
	case "idle", "manual", "ai", "reading", "refactor", "other":
	default:
		return false
	}
	if e.DurationMS > 24*60*60*1000 {
		return false
	}
	return true
}
