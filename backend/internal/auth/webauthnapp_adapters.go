package auth

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/auth/webauthnapp"
)

type webauthnRPAdapter struct{ inner *WebAuthnService }

func (w webauthnRPAdapter) BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	return w.inner.BeginRegistration(ctx, userID)
}

func (w webauthnRPAdapter) FinishRegistration(ctx context.Context, userID uuid.UUID, sid string, body []byte, nickname string) error {
	return w.inner.FinishRegistration(ctx, userID, sid, body, nickname)
}

func (w webauthnRPAdapter) BeginLogin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error) {
	return w.inner.BeginLogin(ctx, email)
}

func (w webauthnRPAdapter) FinishLogin(ctx context.Context, sid string, body []byte) (uuid.UUID, error) {
	return w.inner.FinishLogin(ctx, sid, body)
}

func newWebAuthnApp(w *WebAuthnService) *webauthnapp.Service {
	return webauthnapp.New(webauthnRPAdapter{inner: w})
}
