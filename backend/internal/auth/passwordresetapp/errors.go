package passwordresetapp

import "errors"

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrMissingToken    = errors.New("missing token")
	ErrTokenInvalid    = errors.New("reset token invalid")
)
