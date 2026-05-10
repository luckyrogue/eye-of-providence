package sso

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Registry — кеш OIDCProvider'ов per-team. OIDC discovery (well-known)
// — это HTTP round-trip к IdP'у; не хочется делать его на каждый /sso/start
// call. Cache TTL = 1h: если config меняется через admin UI, мы invalidate'им
// явно через Invalidate(teamID).
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

// NewRegistry — конструктор. redirectURL должен быть public URL endpoint'а
// `/v1/sso/oidc/callback` (registered в IdP'е).
func NewRegistry(pool *pgxpool.Pool, redirectURL string) *Registry {
	return &Registry{
		pool:        pool,
		redirectURL: redirectURL,
		ttl:         1 * time.Hour,
		entries:     make(map[uuid.UUID]*registryEntry),
	}
}

// Get — возвращает provider для team. Кешируется на 1h. Возвращает
// ErrConfigNotFound если SSO не настроен или disabled.
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
		// Phase 1 поддерживает только OIDC. SAML provider будет в отдельном
		// конструкторе.
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

// Invalidate — снимает cache для team'ы (вызывается после SaveConfig в admin
// handler'ах).
func (r *Registry) Invalidate(teamID uuid.UUID) {
	r.mu.Lock()
	delete(r.entries, teamID)
	r.mu.Unlock()
}

// InvalidateAll — full cache flush, для tests или admin reset.
func (r *Registry) InvalidateAll() {
	r.mu.Lock()
	r.entries = make(map[uuid.UUID]*registryEntry)
	r.mu.Unlock()
}
