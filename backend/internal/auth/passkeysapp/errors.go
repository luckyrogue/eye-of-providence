package passkeysapp

import "errors"

var (
	ErrLastAuthFactor  = errors.New("last auth factor")
	ErrPasskeyNotFound = errors.New("passkey not found")
)
