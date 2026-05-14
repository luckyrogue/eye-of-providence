package passwordapp

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrDBNotConfigured    = errors.New("database not configured")
)
