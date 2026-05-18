package passwordresetapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserByEmail interface {
	FindByEmail(ctx context.Context, email string) (id uuid.UUID, locale *string, found bool, err error)
}

type ResetTokenStore interface {
	Insert(ctx context.Context, userID uuid.UUID, tokenHash string, expires time.Time) error
	Consume(ctx context.Context, tokenHash string) (userID uuid.UUID, err error)
}

type ResetMailer interface {
	SendReset(ctx context.Context, to, resetURL string, locale string) error
}

type PasswordSetter interface {
	SetPassword(ctx context.Context, userID uuid.UUID, hash string) error
}

type TokenGenerator interface {
	NewToken() (token, hash string, err error)
	HashToken(token string) string
}
