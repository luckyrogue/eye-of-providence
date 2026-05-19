# Data model

Екі store: **Postgres** transactional state үшін (users, teams, devices,
projects, reports, audit log) және **ClickHouse** analytics үшін (events,
attribution, materialized aggregates). Ingest wire-format шындығы —
[`docs/api/openapi.yaml`](api/openapi.yaml) (`components.schemas.Event`) және
Go типі [`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go).
Транспорт: JSON `POST /v1/ingest` (protobuf/gRPC емес).

Барлық SQL миграциялар `backend/internal/migrate/sql/` версияланады
(`postgres/NNN_*.up.sql` және `clickhouse/NNN_*.up.sql`). API `EOP_AUTO_MIGRATE=true`
болса стартта автоматты қолданады.

## 1. Postgres

### 1.1 Identity & ownership

| Кесте | Мақсат | Ерекше |
|---|---|---|
| `users` | Аккаунттар + soft-delete (`deleted_at`) | `email UNIQUE`, OAuth `github_login`, `team_id` FK, revocation үшін `token_version` |
| `teams` | Пайдаланушы топтары + plan (`free`/`pro`/...) | `settings_json JSONB` икемді team config |
| `team_members` | users↔teams көп-көп + role | `003_multi_team_projects.up.sql` |
| `webauthn_credentials` | FIDO2/passkey credentials | `020_webauthn_credentials.up.sql` |

### 1.2 Devices & projects

| Кесте | Өріс | Неге |
|---|---|---|
| `devices` | `(user_id, fingerprint) UNIQUE` | Desktop agent байлау; pair-flow мұнда row жасайды |
| `projects` | `root_path_hash` (sha256), жол емес | FS ашпай per-project attribution |

### 1.3 Consent & API tokens

- `consent` — scope бойынша granted/revoked timestamps (`telemetry`, `ai_reports`, …).
- `api_tokens` — CI/scripts bearer; `hashed_token TEXT NOT NULL UNIQUE`
  (bcrypt). Plaintext тек бір рет жасалғанда қайтады.

### 1.4 Reports & audit

- `reports` — AI markdown есептер, `period` ∈ `weekly_2026_W18`/`monthly_2026_05`,
  `prompt_version` A/B prompt tracking.
- `audit_log` (`016_audit_log.up.sql`) — security-sensitive
  events: `login_failed`, `password_reset`, `token_created/revoked`,
  `user_deleted`. Retention 24 ай.

### 1.5 Webhooks & rate-limit

- `webhooks` (`011_webhooks.up.sql`) — outbound HTTP (Slack/
  Discord) retry + HMAC secret.
- Rate-limit state — Redis-те, Postgres емес. Failed-attempt audit
  post-mortem үшін `audit_log`.

### 1.6 Cascade & retention

`users(id)` FK-лерінде `ON DELETE CASCADE` — `DELETE /v1/me/data`
бір `DELETE FROM users WHERE id = $1` арқылы profile + devices + projects + consent + reports +
api_tokens өшіреді. Soft-delete
(`users.deleted_at`) тек login-blocking, retention threshold емес.

## 2. ClickHouse

### 2.1 Raw events — `events`

```
events (ts, user_id, device_id, session_id,
        app_bundle, category, source, ai_provider, ai_channel,
        project_id, file_lang,
        duration_ms, chars_in, lines_added, lines_removed,
        meta)
ORDER BY (user_id, ts)
PARTITION BY toYYYYMM(ts)
TTL toDateTime(ts) + INTERVAL 18 MONTH
```

- `category` ∈ `idle`, `manual`, `ai`, `reading`, `refactor`, `other`.
- `source` ∈ `os`, `browser`, `ide`, `cli`.
- `meta String` — JSON қауіпсіз кеңейтулер: `clipboard_sha256`,
  `clipboard_bytes`, `mouse_clicks`. Мазмұн жоқ ([/privacy](/privacy) §1.1).
- TTL 18 ай — partition бойынша auto-drop. Retention үшін `DELETE WHERE` қолданбаймыз
  (қымбат); тек `DELETE /v1/me/data`.

### 2.2 Attribution — `attribution_events`

`events`-тен `attribution worker` арқылы (`backend/internal/attribution/`).
Мұндағы категориялар raw-дан нақты:

```
category Enum8(
    'typed'        = 1,
    'pasted_ai'    = 2,
    'pasted_other' = 3,
    'ai_inline'    = 4,
    'ai_agent'     = 5,
    'refactor'     = 6,
    'unknown'      = 7
)
```

Worker batch-тарда жүреді. Фазалар
[`backend/internal/attribution/worker.go`](../backend/internal/attribution/worker.go):

**Phase A (қазір):**
1. `category in (manual, ai)` жаңа `events` оқиды;
2. әр event-ті flat ережелермен `attribution_events.category`-ға маппингтейді;
3. `attribution_events`-ке batch жазады, `(ts, user_id, project_id)` dedup.

**Phase B (roadmap):**
1. paste `clipboard_sha256` + `clipboard_signal` join — 30 сек → `pasted_ai` vs `pasted_other`;
2. IDE extension diff snapshots арқылы per-hunk;
3. direct API hooks бар жерде heuristicсіз inline/agent.

### 2.3 Materialized aggregates

Cascading MVs (`backend/internal/migrate/sql/clickhouse/003_*`, `004_*`):

```
events (raw)  →  events_hourly_agg  →  events_daily_agg
              events_hourly_mv       events_daily_mv
```

- `events_hourly_agg` — `SummingMergeTree` `(user_id, bucket_ts, category, file_lang)`.
- `events_daily_agg` — ұзақ диапазон (30/90 күн).

**Routing** (`store/clickhouse.go`): `since ≥ 30d` → **daily** MV, әйтпесе hourly.

**Reduction** (100 active users): raw ~10M/күн → hourly ~290K (**35×**) → daily ~12K (**840×**).

**Benchmark** (`backend/` ішінен `go run ./cmd/ch-bench`):

```sh
EOP_CH_DSN="clickhouse://default:@localhost:19000/eop_test" \
  go run ./cmd/ch-bench --users=10 --days=30 --events-per-user-per-day=10000
```

### 2.4 Redis read cache (`store/cached.go`)

Redis қолжетімді болса `EventStore` decorator. TTL: `AggregateByCategory` 10m,
`AggregateByCategoryBulk` 5m, т.б. Redis down → miss + log, fail емес.

## 3. Event flow

```
┌──────────┐  events     ┌──────────┐   raw insert    ┌────────────┐
│  Agent   │ ──────────► │  /v1/    │ ──────────────► │ events     │
│ (desktop,│  batched    │ ingest   │   ClickHouse    │ MergeTree  │
│ browser, │  HTTP+JSON  │ handler  │                 └─────┬──────┘
│ IDE, CLI)│             │ (Go)     │                       │ MV trigger
└──────────┘             └──────────┘                       ▼
                                                    ┌────────────────────┐
                                                    │ events_hourly_agg  │
                                                    │ events_daily_agg   │
                                                    └────────────────────┘

                                ┌─────────────────┐  batch SELECT
                                │ attribution-    │ ────────────► events
                                │ worker (Go)     │
                                └────────┬────────┘
                                         ▼
                                 ┌────────────────────┐
                                 │ attribution_events │
                                 └────────────────────┘
```

## 4. Cross-cutting invariants

- **`user_id` әрқашан UUIDv4**, email немесе client-supplied емес. Backend JWT-claims-тен
  қайта жазады (`ingestapp/service.go`).
- **Мазмұн жоқ**. Тек санағыштар, hash, enum. `meta` — қауіпсіз кілттер ғана.
  Толығырақ — [/privacy](/privacy) §1.1.
- **Idempotent migrations**. `001..N` advisory_lock + `IF NOT EXISTS`.
- **TTL — analytics retention-ның жалғыз механизмі**. GDPR erasure — explicit
  `ALTER TABLE ... DELETE WHERE user_id = ?`.

## 5. Кеңейту

Event-ке жаңа өріс қосқанда:
1. [`docs/api/openapi.yaml`](api/openapi.yaml) `components.schemas.Event`.
2. `backend/internal/ingest/domain/event.go` және `backend/internal/store/event.go`.
3. `005_*.up.sql` миграция.
4. `clickhouse.go` `Insert()`.
5. Агрегация керек болса — жаңа MV.
6. Agent: `agent/src-tauri/src/core/event.rs`.

Мазмұнсыз кеңейтулер үшін `meta map<string,string>` (миграция қажет емес).
