package sso

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Registry struct {
	pool        *pgxpool.Pool
	redirectURL string
	ttl         time.Duration
	mu          sync.RWMutex
	entries     map[uuid.UUID]*registryEntry
}

type registryEntry struct {
	provider *OIDCProvider
	loadedAt time.Time
}

func NewRegistry(pool *pgxpool.Pool, redirectURL string) *Registry {
	return &Registry{
		pool:        pool,
		redirectURL: redirectURL,
		ttl:         1 * time.Hour,
		entries:     make(map[uuid.UUID]*registryEntry),
	}
}

func (r *Registry) Get(ctx context.Context, teamID uuid.UUID) (*OIDCProvider, error) {
	r.mu.RLock()
	if e, ok := r.entries[teamID]; ok && time.Since(e.loadedAt) < r.ttl {
		r.mu.RUnlock()
		return e.provider, nil
	}
	r.mu.RUnlock()

	cfg, err := LoadConfig(ctx, r.pool, teamID)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, ErrConfigNotFound
	}
	if cfg.Provider != ProviderOIDC {

		return nil, errors.New("non-OIDC provider not supported in Phase 1")
	}

	prov, err := NewOIDCProvider(ctx, cfg, r.redirectURL)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.entries[teamID] = &registryEntry{provider: prov, loadedAt: time.Now()}
	r.mu.Unlock()
	return prov, nil
}

func (r *Registry) Invalidate(teamID uuid.UUID) {
	r.mu.Lock()
	delete(r.entries, teamID)
	r.mu.Unlock()
}

func (r *Registry) InvalidateAll() {
	r.mu.Lock()
	r.entries = make(map[uuid.UUID]*registryEntry)
	r.mu.Unlock()
}
