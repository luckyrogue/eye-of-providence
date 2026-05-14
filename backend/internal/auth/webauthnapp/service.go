package webauthnapp

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

// RP — операции WebAuthn relying party (реализация: *auth.WebAuthnService).
type RP interface {
	BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error)
	FinishRegistration(ctx context.Context, userID uuid.UUID, sid string, body []byte, nickname string) error
	BeginLogin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error)
	FinishLogin(ctx context.Context, sid string, body []byte) (uuid.UUID, error)
}

// Service — use case-слой над RP (тонкие Fiber-handlers делегируют сюда).
type Service struct {
	rp RP
}

// New — конструктор.
func New(rp RP) *Service {
	return &Service{rp: rp}
}

// RegisterBegin — POST /v1/auth/webauthn/register/begin.
func (s *Service) RegisterBegin(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	return s.rp.BeginRegistration(ctx, userID)
}

// RegisterFinish — POST /v1/auth/webauthn/register/finish.
func (s *Service) RegisterFinish(ctx context.Context, userID uuid.UUID, sessionID string, attestation []byte, nickname string) error {
	return s.rp.FinishRegistration(ctx, userID, sessionID, attestation, nickname)
}

// LoginBegin — POST /v1/auth/webauthn/login/begin.
func (s *Service) LoginBegin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error) {
	return s.rp.BeginLogin(ctx, email)
}

// LoginFinish — POST /v1/auth/webauthn/login/finish.
func (s *Service) LoginFinish(ctx context.Context, sessionID string, assertion []byte) (uuid.UUID, error) {
	return s.rp.FinishLogin(ctx, sessionID, assertion)
}
