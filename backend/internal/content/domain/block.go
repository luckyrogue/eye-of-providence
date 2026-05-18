package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Block — aggregate root CMS-блока (slug + locale).
type Block struct {
	Slug          string
	Locale        string
	Content       json.RawMessage
	DraftContent  json.RawMessage
	SchemaVersion int
	PublishedAt   *time.Time
	UpdatedAt     time.Time
	UpdatedBy     *uuid.UUID
}

type MatrixEntry struct {
	Slug         string
	Locale       string
	HasPublished bool
	HasDraft     bool
	UpdatedAt    *time.Time
	UpdatedBy    *uuid.UUID
}

type UpsertParams struct {
	Slug           string
	Locale         string
	Content        json.RawMessage
	Publish        bool
	SchemaVersion  int
	UpdatedBy      uuid.UUID
	PriorUpdatedAt *time.Time
}
