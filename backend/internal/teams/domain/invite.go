package domain

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID       uuid.UUID
	TeamID   uuid.UUID
	Code     string
	MaxUses  int
	UseCount int
	Expires  *time.Time
}
