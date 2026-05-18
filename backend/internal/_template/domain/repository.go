package domain

import "context"

// Repository — persistence port owned by the domain layer.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Entity, error)
}
