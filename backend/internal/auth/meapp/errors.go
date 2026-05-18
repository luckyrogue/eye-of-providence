package meapp

import "errors"

var (
	ErrInvalidSubject    = errors.New("invalid token subject")
	ErrInvalidBody       = errors.New("invalid body")
	ErrUnsupportedLocale = errors.New("unsupported locale")
	ErrInvalidTokenID    = errors.New("invalid token id")
	ErrNameTooLong       = errors.New("name too long (max 64)")
	ErrTTLOutOfRange     = errors.New("ttl_days must be 0..365")
	ErrTokenNotFound     = errors.New("token not found")
	ErrJWTRequired       = errors.New("tokens management requires JWT")
	ErrDBNotConfigured   = errors.New("database not configured")
	ErrNoFields          = errors.New("no fields")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidLastName   = errors.New("invalid last name")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrNoPasswordSet     = errors.New("no password set")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken        = errors.New("email taken")
	ErrInvalidPassword   = errors.New("invalid password")
)
