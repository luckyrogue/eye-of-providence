package adminapp

import "errors"

var (
	ErrSSONotConfigured    = errors.New("sso not configured")
	ErrCannotDeleteSelf    = errors.New("cannot delete yourself")
	ErrLastSuperAdmin      = errors.New("cannot delete last super_admin")
	ErrCannotDemoteSelf    = errors.New("cannot demote yourself")
	ErrInvalidGlobalRole   = errors.New("global_role must be user or super_admin")
	ErrInvalidDisplayName  = errors.New("invalid display_name")
	ErrUserNotFound        = errors.New("user not found")
	ErrOwnerLimit          = errors.New("user already owns another company")
	ErrInvalidRole         = errors.New("role must be owner, admin, or member")
	ErrInvalidEmail        = errors.New("valid email required")
	ErrInvalidPlan         = errors.New("invalid subscription plan")
	ErrInvalidUntil        = errors.New("until must be ISO8601")
	ErrInvalidPayment      = errors.New("invalid payment")
)
