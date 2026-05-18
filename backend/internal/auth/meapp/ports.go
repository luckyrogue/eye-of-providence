package meapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionClaims struct {
	UserID   string
	Email    string
	Provider string
}

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

type ProfileReader interface {
	LoadExtras(ctx context.Context, userID uuid.UUID) (*ProfileExtras, error)
	UpdateLocale(ctx context.Context, userID uuid.UUID, locale string) error
}

type ProfileWriter interface {
	UpdateName(ctx context.Context, userID uuid.UUID, displayName, lastName *string) error
	PasswordHash(ctx context.Context, userID uuid.UUID) (hash string, hasPassword bool, err error)
	UpdateEmail(ctx context.Context, userID uuid.UUID, email, password string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error
}

type SessionIssuer interface {
	IssueAfterCredentialChange(ctx context.Context, userID uuid.UUID, email string) (token string, err error)
}

type TokenRow struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Scope      string     `json:"scope"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type TokenWriter interface {
	List(ctx context.Context, userID uuid.UUID) ([]TokenRow, error)
	Create(ctx context.Context, userID uuid.UUID, name, scope string, ttl time.Duration) (plaintext string, meta TokenRow, err error)
	Revoke(ctx context.Context, userID, tokenID uuid.UUID) (bool, error)
}
