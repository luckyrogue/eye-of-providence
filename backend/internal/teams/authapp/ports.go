package authapp

import (
	"context"

	"github.com/google/uuid"
)

type LoginUser struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
}

type PasswordAuthenticator interface {
	VerifyLogin(ctx context.Context, email, password string) (LoginUser, error)
}
