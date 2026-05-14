package meapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionClaims — идентичность из JWT (без Fiber).
type SessionClaims struct {
	UserID   string
	Email    string
	Provider string
}

// ProfileExtras — поля из users row (опционально).
type ProfileExtras struct {
	GithubLogin  *string
	GlobalRole   *string
	DisplayName  *string
	LastName     *string
	Phone        *string
	Locale       *string
	HasPassword  bool
	CreatedAtRFC string
}

// ProfileReader — догрузка профиля из БД.
type ProfileReader interface {
	LoadExtras(ctx context.Context, userID uuid.UUID) (*ProfileExtras, error)
	UpdateLocale(ctx context.Context, userID uuid.UUID, locale string) error
}

// TokenRow — метаданные API token (без plaintext).
type TokenRow struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// TokenWriter — CRUD API tokens.
type TokenWriter interface {
	List(ctx context.Context, userID uuid.UUID) ([]TokenRow, error)
	Create(ctx context.Context, userID uuid.UUID, name, scope string, ttl time.Duration) (plaintext string, meta TokenRow, err error)
	Revoke(ctx context.Context, userID, tokenID uuid.UUID) (bool, error)
}
