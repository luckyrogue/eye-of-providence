package emailtemplates

import (
	"time"

	"github.com/google/uuid"
)

type MatrixEntry struct {
	Key         string     `json:"key"`
	Locale      string     `json:"locale"`
	HasOverride bool       `json:"has_override"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

type View struct {
	Key       string     `json:"key"`
	Locale    string     `json:"locale"`
	Subject   string     `json:"subject"`
	BodyHTML  string     `json:"body_html"`
	BodyText  string     `json:"body_text"`
	IsDefault bool       `json:"is_default"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

type UpsertCommand struct {
	Subject  string
	BodyHTML string
	BodyText string
}

type UpsertResult struct {
	Key       string     `json:"key"`
	Locale    string     `json:"locale"`
	Subject   string     `json:"subject"`
	BodyHTML  string     `json:"body_html"`
	BodyText  string     `json:"body_text"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *uuid.UUID `json:"updated_by"`
	IsDefault bool       `json:"is_default"`
}

type OverrideRow struct {
	Key       string
	Locale    string
	Subject   string
	BodyHTML  string
	BodyText  string
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
}

type AuditEvent struct {
	ActorID    uuid.UUID
	ActorEmail string
	Action     string
	TargetType string
	TargetID   string
	Metadata   map[string]any
	IP         string
	UserAgent  string
}

type RequestMeta struct {
	IP        string
	UserAgent string
}
