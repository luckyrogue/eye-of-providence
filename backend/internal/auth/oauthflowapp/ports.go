package oauthflowapp

import (
	"context"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/oauthapp"
)

type Provider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (oauthapp.ExternalUser, error)
}

type UserLinker interface {
	UpsertOAuthUser(ctx context.Context, provider string, ext oauthapp.ExternalUser) (uuid.UUID, error)
}

type SessionIssuer interface {
	IssueHandoff(ctx context.Context, userID uuid.UUID, email, method string) (string, error)
}
