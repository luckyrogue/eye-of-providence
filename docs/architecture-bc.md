# Backend bounded contexts (Clean Architecture + DDD)

## Layer flow

```
HTTP (handler.go) → *app.Service → domain (entities, rules) ← infrastructure (store, *_adapters.go)
```

Dependencies point inward only. `domain/` must not import `fiber`, `pgx`, `zap`, or other product BCs.

## Product bounded contexts

| BC | Aggregate roots (primary) | Notes |
|----|-------------------------|--------|
| content | Block (slug+locale) | Reference implementation |
| teams | Team, TeamMember, TeamInvite, Project, Commit | Largest legacy surface |
| auth | User, Identity, Passkey, APIToken | OAuth/WebAuthn in application |
| ingest | Event (batch) | Write path to ClickHouse |
| analytics | Event (read models) | Dashboard queries |
| publicapi | Event (read, scoped) | API token read scope |
| insights | Insight | Generated from events |
| reports | Report | PG + optional Gemini |
| webhooks | Webhook | HMAC delivery |
| push | PushSubscription | Web Push |
| devices | Device, PairingSession | Pair/claim flow |
| sso | SSOConfig | OIDC team SSO |
| prcomment | CommentBody | PR comment formatting |
| anomaly | AnomalySignal | Detector + cron |
| attribution | CommitAttribution | Worker only |

## Shared kernel (not BCs)

`httperr`, `log`, `config`, `migrate`, `metrics`, `audit`, `cache`, `plans`, `mailer`, `store` (implementations).

BCs declare **narrow ports** (e.g. `EventReadStore` with only methods they need). Adapters wrap `store.EventStore` in `*_adapters.go`.

| Port | Consumers | Methods (typical) |
|------|-----------|-------------------|
| `EventReadStore` | analytics, publicapi, insights, reports | ListRecent, AggregateByCategory, LanguageBreakdown, DailyTrend, Heatmap, ActiveUserIDs |
| `EventWriteStore` | ingest | Insert |

Handlers receive `store.EventStore` only at composition root ([`backend/cmd/api/main.go`](../backend/cmd/api/main.go), [`backend/internal/app`](../backend/internal/app/modules.go)); BC adapters map to domain types.

## Cross-BC rules

1. No import of another BC's `handler.go` or Fiber handlers.
2. Call other BCs via **application ports** implemented in `*_adapters.go` (e.g. `teams` → `PasswordAuthenticator` port → `auth/passwordapp`).
3. Do not share `domain/` packages across BCs; duplicate read DTOs if needed.
4. Stable `httperr` codes and JSON response shapes are part of the public contract — do not change during refactors.

## New slice checklist

See [backend/AGENTS.md](../backend/AGENTS.md) section «Чеклист нового bounded context». Migration status by domain — same file, table «Статус миграции».

## Template

Copy from [`backend/internal/_template/`](../backend/internal/_template/).
