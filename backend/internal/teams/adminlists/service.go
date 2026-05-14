package adminlists

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WebhookRow — cross-team webhook list row.
type WebhookRow struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	UserEmail      string     `json:"user_email"`
	URL            string     `json:"url"`
	Events         []string   `json:"events"`
	Format         string     `json:"format"`
	Active         bool       `json:"active"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     *int       `json:"last_status,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// APITokenRow — cross-user API token list row.
type APITokenRow struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	UserEmail  string     `json:"user_email"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

// ListQuerier — read-only admin list queries.
type ListQuerier interface {
	ListWebhooks(ctx context.Context, limit, offset int) ([]WebhookRow, error)
	ListAPITokens(ctx context.Context, limit, offset int, includeRevoked bool) ([]APITokenRow, error)
}

// Service — admin read models.
type Service struct {
	q ListQuerier
}

// New — wiring.
func New(q ListQuerier) *Service {
	return &Service{q: q}
}

// ListWebhooks — paginated.
func (s *Service) ListWebhooks(ctx context.Context, limit, offset int) ([]WebhookRow, error) {
	if limit > 100 {
		limit = 100
	}
	return s.q.ListWebhooks(ctx, limit, offset)
}

// ListAPITokens — paginated.
func (s *Service) ListAPITokens(ctx context.Context, limit, offset int, includeRevoked bool) ([]APITokenRow, error) {
	if limit > 100 {
		limit = 100
	}
	return s.q.ListAPITokens(ctx, limit, offset, includeRevoked)
}
