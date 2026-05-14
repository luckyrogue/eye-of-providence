package teamflags

import "errors"

var (
	ErrTeamNotFound = errors.New("team not found")
	ErrMissingFlags = errors.New("flags object required")
)
