package ssoapp

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	reg    Registry
	states StateCreator
}

type Deps struct {
	Registry Registry
	States   StateCreator
}

func New(d Deps) *Service {
	return &Service{reg: d.Registry, states: d.States}
}

func (s *Service) AuthorizeURL(ctx context.Context, teamID uuid.UUID, returnTo string) (string, error) {
	if s.reg == nil || s.states == nil {
		return "", nil
	}
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
