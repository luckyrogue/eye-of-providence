package passkeysapp

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	rp      PasskeyRP
	factors AuthFactorCounter
}

type Deps struct {
	RP      PasskeyRP
	Factors AuthFactorCounter
}

func New(d Deps) *Service {
	return &Service{rp: d.RP, factors: d.Factors}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]PasskeyRow, error) {
	if s.rp == nil {
		return nil, nil
	}
	return s.rp.ListPasskeys(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID, passkeyID uuid.UUID) error {
	if s.rp == nil {
		return ErrPasskeyNotFound
	}
	credID, err := s.rp.PasskeyCredentialIDForUser(ctx, userID, passkeyID)
	if err != nil {
		if errors.Is(err, ErrPasskeyNotFound) {
			return ErrPasskeyNotFound
		}
		return err
	}
	if s.factors != nil {
		n, err := s.factors.Count(ctx, userID, nil, credID)
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrLastAuthFactor
		}
	}
	if err := s.rp.DeletePasskey(ctx, userID, passkeyID); err != nil {
		if errors.Is(err, ErrPasskeyNotFound) {
			return ErrPasskeyNotFound
		}
		return err
	}
	return nil
}
