package apitokensapp

import "context"

// TokenStore — port for API token persistence (implemented in delivery adapters).
type TokenStore interface {
	List(ctx context.Context, userID string) ([]TokenRow, error)
}

type TokenRow struct {
	ID     string `json:"id"`
	Prefix string `json:"prefix"`
	Label  string `json:"label,omitempty"`
}

type Service struct {
	store TokenStore
}

func New(store TokenStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context, userID string) ([]TokenRow, error) {
	if s.store == nil {
		return []TokenRow{}, nil
	}
	return s.store.List(ctx, userID)
}
