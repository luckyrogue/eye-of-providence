package ssostart

import (
	"context"

	"github.com/google/uuid"
)

// Service — POST /v1/sso/start.
type Service struct {
	reg    Registry
	states StateCreator
}

// New — конструктор.
func New(reg Registry, states StateCreator) *Service {
	return &Service{reg: reg, states: states}
}

// AuthorizeURL — IdP authorization URL для SPA redirect.
func (s *Service) AuthorizeURL(ctx context.Context, teamID uuid.UUID, returnTo string) (string, error) {
	prov, err := s.reg.GetOIDC(ctx, teamID)
	if err != nil {
		return "", err
	}
	sv, nonce, err := s.states.CreateState(ctx, teamID, returnTo)
	if err != nil {
		return "", err
	}
	return prov.AuthCodeURL(sv, nonce), nil
}
