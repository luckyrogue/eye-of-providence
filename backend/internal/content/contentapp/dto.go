package contentapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const MaxContentBytes = 256 * 1024

type PublicView struct {
	Slug          string
	Locale        string
	Content       json.RawMessage
	SchemaVersion int
	PublishedAt   *time.Time
	UpdatedAt     time.Time
	ETag          string
	Source        string
	CacheHit      bool
	NotModified   bool
}

type PreviewView struct {
	PublicView
}

type AdminMatrixView struct {
	Entries []MatrixEntryView
}

type MatrixEntryView struct {
	Slug         string
	Locale       string
	HasPublished bool
	HasDraft     bool
	UpdatedAt    *time.Time
	UpdatedBy    *uuid.UUID
}

type AdminBlockView struct {
	Slug          string
	Locale        string
	Content       json.RawMessage
	DraftContent  json.RawMessage
	SchemaVersion int
	PublishedAt   *time.Time
	UpdatedAt     *time.Time
	UpdatedBy     *uuid.UUID
	ETag          string
	Empty         bool
}

type UpsertCommand struct {
	Slug           string
	Locale         string
	Content        json.RawMessage
	Publish        bool
	ActorID        uuid.UUID
	ActorEmail     string
	PriorUpdatedAt *time.Time
}

type UpsertResult struct {
	AdminBlockView
	Published bool
}

type DeleteResult struct {
	Reverted bool
	Prior    AdminBlockView
}

type SchemaViolation struct {
	Code       string
	Field      string
	Detail     string
	SchemaPath string
}

func (e *SchemaViolation) Error() string { return e.Detail }
