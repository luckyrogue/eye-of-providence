package store

import (
	"context"

	"github.com/google/uuid"
)

// DeleteUserData — добавочный contract: удаление всех событий пользователя.
// Не часть EventStore интерфейса (опциональная capability).
type UserDeleter interface {
	DeleteUserData(ctx context.Context, userID string) error
}

func (s *MemoryStore) DeleteUserData(_ context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.events[:0]
	for _, e := range s.events {
		if e.UserID != userID {
			out = append(out, e)
		}
	}
	s.events = out
	return nil
}

func (s *ClickHouseStore) DeleteUserData(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	// ALTER TABLE … DELETE — асинхронная mutation, но мы не ждём materializ-ции.
	return s.conn.Exec(ctx, "ALTER TABLE events DELETE WHERE user_id = ?", uid)
}
