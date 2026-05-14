package webauthnapp

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

type RP interface {
	BeginRegistration(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error)
	FinishRegistration(ctx context.Context, userID uuid.UUID, sid string, body []byte, nickname string) error
	BeginLogin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error)
	FinishLogin(ctx context.Context, sid string, body []byte) (uuid.UUID, error)
}

type Service struct {
	rp RP
}

func New(rp RP) *Service {
	return &Service{rp: rp}
}

func (s *Service) RegisterBegin(ctx context.Context, userID uuid.UUID) (*protocol.CredentialCreation, string, error) {
	return s.rp.BeginRegistration(ctx, userID)
}

func (s *Service) RegisterFinish(ctx context.Context, userID uuid.UUID, sessionID string, attestation []byte, nickname string) error {
	return s.rp.FinishRegistration(ctx, userID, sessionID, attestation, nickname)
}

func (s *Service) LoginBegin(ctx context.Context, email *string) (*protocol.CredentialAssertion, string, error) {
	return s.rp.BeginLogin(ctx, email)
}

func (s *Service) LoginFinish(ctx context.Context, sessionID string, assertion []byte) (uuid.UUID, error) {
	return s.rp.FinishLogin(ctx, sessionID, assertion)
}
