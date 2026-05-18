package identitiesapp

import "errors"

var (
	ErrLastAuthFactor = errors.New("last auth factor")
	ErrNotFound       = errors.New("not found")
	ErrDBRequired     = errors.New("database required")
)
