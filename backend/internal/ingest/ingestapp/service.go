package ingestapp

import (
	"context"
	"errors"
	"time"

	"github.com/eye-of-providence/backend/internal/ingest/domain"
)

type Service struct {
	writer EventWriter
}

type Deps struct {
	Writer EventWriter
}

func New(d Deps) *Service {
	return &Service{writer: d.Writer}
}

type BatchResult struct {
	Accepted int
	Rejected int
}

func (s *Service) PrepareBatch(userID string, events []domain.Event, maxBatch int) ([]domain.Event, BatchResult, error) {
	if len(events) > maxBatch {
		return nil, BatchResult{}, domain.ErrBatchTooLarge
	}
	now := time.Now().UTC()
	valid := make([]domain.Event, 0, len(events))
	accepted, rejected := 0, 0
	for _, e := range events {
		if !domain.ValidEvent(e) {
			rejected++
			continue
		}
		e.UserID = userID
		if e.TS.IsZero() {
			e.TS = now
		}
		valid = append(valid, e)
		accepted++
	}
	return valid, BatchResult{Accepted: accepted, Rejected: rejected}, nil
}

func (s *Service) PersistBatch(ctx context.Context, events []domain.Event) error {
	if len(events) == 0 || s.writer == nil {
		return nil
	}
	return s.writer.Insert(ctx, events)
}

var ErrBatchTooLarge = domain.ErrBatchTooLarge

func IsBatchTooLarge(err error) bool {
	return errors.Is(err, domain.ErrBatchTooLarge)
}
