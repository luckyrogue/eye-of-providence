package meapp

import "errors"

var (
	ErrInvalidSubject   = errors.New("invalid token subject")
	ErrInvalidBody      = errors.New("invalid body")
	ErrUnsupportedLocale = errors.New("unsupported locale")
	ErrInvalidTokenID   = errors.New("invalid token id")
	ErrNameTooLong      = errors.New("name too long (max 64)")
	ErrTTLOutOfRange    = errors.New("ttl_days must be 0..365")
	ErrTokenNotFound    = errors.New("token not found")
	ErrJWTRequired        = errors.New("tokens management requires JWT")
	ErrDBNotConfigured    = errors.New("database not configured")
)
