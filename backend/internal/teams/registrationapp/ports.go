package registrationapp

import "context"

type UserCounter interface {
	UserCount(ctx context.Context) (int, error)
}

type FirstUserPromoter interface {
	PromoteSuperAdmin(ctx context.Context, userID string) error
}
