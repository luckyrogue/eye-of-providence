package emailtemplates

import "errors"

// Sentinel errors returned by application service (handlers map to httperr).
var (
	ErrInvalidKey            = errors.New("invalid template key")
	ErrInvalidLocale         = errors.New("invalid template locale")
	ErrMissingField          = errors.New("subject and body_html are required")
	ErrBodyTooLarge          = errors.New("body exceeds 256 KB")
	ErrStoreUnavailable      = errors.New("template override store not configured")
	ErrNoBaseline            = errors.New("no baseline for key and locale")
	ErrNoOverride            = errors.New("no override exists for key and locale")
)
