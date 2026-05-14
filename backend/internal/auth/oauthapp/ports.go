package oauthapp

import (
	"context"

	"github.com/google/uuid"
)

// Store — персистентность для upsert OAuth identity (реализация в родительском пакете auth).
type Store interface {
	FindUserIDByIdentity(ctx context.Context, provider, subject string) (userID uuid.UUID, ok bool, err error)
	UpdateUserEmailIfEmpty(ctx context.Context, userID uuid.UUID, email string) error
	FindUserIDByEmail(ctx context.Context, email string) (userID uuid.UUID, ok bool, err error)
	LinkIdentity(ctx context.Context, userID uuid.UUID, provider, subject, email string) error
	CreateUserWithIdentity(ctx context.Context, newID uuid.UUID, provider string, ext ExternalUser) error
}
