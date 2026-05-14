package ssostart

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	reg    Registry
	states StateCreator
}

func New(reg Registry, states StateCreator) *Service {
	return &Service{reg: reg, states: states}
}

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
