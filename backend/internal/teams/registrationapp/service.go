package registrationapp

import "context"

type Service struct {
	users UserCounter
	promo FirstUserPromoter
}

type Deps struct {
	Users UserCounter
	Promo FirstUserPromoter
}

func New(d Deps) *Service {
	return &Service{users: d.Users, promo: d.Promo}
}

type RegisterContext struct {
	IsFirstUser bool
}

func (s *Service) BeforeRegister(ctx context.Context) (RegisterContext, error) {
	if s.users == nil {
		return RegisterContext{}, nil
	}
	n, err := s.users.UserCount(ctx)
	if err != nil {
		return RegisterContext{}, err
	}
	return RegisterContext{IsFirstUser: n == 0}, nil
}

func (s *Service) AfterRegister(ctx context.Context, userID string, rc RegisterContext) error {
	if !rc.IsFirstUser || s.promo == nil {
		return nil
	}
	return s.promo.PromoteSuperAdmin(ctx, userID)
}
