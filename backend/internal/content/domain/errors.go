package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrUnavailable = errors.New("content store unavailable: pool nil")
	ErrNotFound    = errors.New("content block not found")
)

type ErrPrecondition struct {
	CurrentUpdatedAt time.Time
}

func (e *ErrPrecondition) Error() string {
	return fmt.Sprintf("if-match precondition failed: current updated_at=%s",
		e.CurrentUpdatedAt.UTC().Format(time.RFC3339Nano))
}
