package oauthapp

import (
	"context"

	"github.com/google/uuid"
)

// Service — upsert пользователя после OAuth callback.
type Service struct {
	store Store
}

// New — конструктор. store == nil → in-memory режим (детерминированный UUID).
func New(store Store) *Service {
	return &Service{store: store}
}

// UpsertOAuthUser — шаги 1–3 из прежнего handler.go: identity → email → create.
func (s *Service) UpsertOAuthUser(ctx context.Context, provider string, ext ExternalUser) (uuid.UUID, error) {
	if s.store == nil {
		return uuid.NewSHA1(uuid.NameSpaceURL, []byte(provider+":"+ext.Subject)), nil
	}

	if uid, ok, err := s.store.FindUserIDByIdentity(ctx, provider, ext.Subject); err != nil {
		return uuid.Nil, err
	} else if ok {
		_ = s.store.UpdateUserEmailIfEmpty(ctx, uid, ext.Email)
		return uid, nil
	}

	if uid, ok, err := s.store.FindUserIDByEmail(ctx, ext.Email); err != nil {
		return uuid.Nil, err
	} else if ok {
		if err := s.store.LinkIdentity(ctx, uid, provider, ext.Subject, ext.Email); err != nil {
			return uuid.Nil, err
		}
		return uid, nil
	}

	newID := uuid.New()
	if err := s.store.CreateUserWithIdentity(ctx, newID, provider, ext); err != nil {
		return uuid.Nil, err
	}
	return newID, nil
}
