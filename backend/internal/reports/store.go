package reports

import (
	"sort"
	"sync"
	"time"
)

type Report struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Period        string    `json:"period"`
	Model         string    `json:"model"`
	BodyMD        string    `json:"body_md"`
	GeneratedAt   time.Time `json:"generated_at"`
	PromptVersion string    `json:"prompt_version"`
}

// ReportStore — общий интерфейс для in-memory и Postgres реализаций.
type ReportStore interface {
	Save(r Report)
	ListForUser(userID string, limit int) []Report
	Get(id, userID string) (Report, bool)
}

// Store — in-memory реализация ReportStore.
type Store struct {
	mu      sync.RWMutex
	reports []Report
}

func NewStore() *Store {
	return &Store{reports: make([]Report, 0, 32)}
}

func (s *Store) Save(r Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports = append(s.reports, r)
}

func (s *Store) ListForUser(userID string, limit int) []Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Report, 0, limit)
	for i := len(s.reports) - 1; i >= 0 && len(out) < limit; i-- {
		if s.reports[i].UserID == userID {
			out = append(out, s.reports[i])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].GeneratedAt.After(out[j].GeneratedAt) })
	return out
}

func (s *Store) Get(id, userID string) (Report, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reports {
		if r.ID == id && r.UserID == userID {
			return r, true
		}
	}
	return Report{}, false
}
