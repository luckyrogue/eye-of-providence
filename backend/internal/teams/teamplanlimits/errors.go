package teamplanlimits

import "errors"

var (
	ErrTeamNotFound  = errors.New("team not found")
	ErrMissingLimits = errors.New("limits field required (use null to reset)")
)
