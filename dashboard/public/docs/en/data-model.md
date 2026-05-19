# Data model

Two stores: **Postgres** for transactional state (users, teams, devices,
projects, reports, audit log) and **ClickHouse** for analytics (events,
attribution, materialized aggregates). Source of truth for the ingest wire format —
[`docs/api/openapi.yaml`](api/openapi.yaml) (`components.schemas.Event`) and
Go type [`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go).
Transport: JSON `POST /v1/ingest` (not protobuf/gRPC).

All SQL migrations are version-controlled in `backend/internal/migrate/sql/`
(`postgres/NNN_*.up.sql` and `clickhouse/NNN_*.up.sql`). The API applies them
automatically on start if `EOP_AUTO_MIGRATE=true`.

## 1. Postgres

### 1.1 Identity & ownership

| Table | Purpose | Notable |
|---|---|---|
| `users` | Accounts + soft-delete (`deleted_at`) | `email UNIQUE`, OAuth `github_login`, `team_id` FK, `token_version` for revocation |
| `teams` | User groups + plan (`free`/`pro`/...) | `settings_json JSONB` for flexible per-team config |
| `team_members` | Many-to-many users↔teams + role | See `003_multi_team_projects.up.sql` |
| `webauthn_credentials` | FIDO2/passkey credentials | See `020_webauthn_credentials.up.sql` |

### 1.2 Devices & projects

| Table | Field | Why |
|---|---|---|
| `devices` | `(user_id, fingerprint) UNIQUE` | Desktop agent binding; pair-flow creates row here |
| `projects` | `root_path_hash` (sha256), not the path itself | Per-project attribution without FS disclosure |

### 1.3 Consent & API tokens

- `consent` — granted/revoked timestamps by scope (`telemetry`, `ai_reports`, …).
- `api_tokens` — bearer tokens for CI/scripts; `hashed_token TEXT NOT NULL UNIQUE`
  (bcrypt). Plaintext returned exactly once on creation.

### 1.4 Reports & audit

- `reports` — AI-generated markdown reports, `period` ∈ `weekly_2026_W18`/`monthly_2026_05`,
  `prompt_version` for A/B prompt tracking.
- `audit_log` (migration `016_audit_log.up.sql`) — security-sensitive
  events: `login_failed`, `password_reset`, `token_created/revoked`,
  `user_deleted`. Retention 24 months.

### 1.5 Webhooks & rate-limit

- `webhooks` (`011_webhooks.up.sql`) — outbound HTTP notifications (Slack/
  Discord) with retry state and HMAC secret.
- Rate-limit state — in Redis, not Postgres. Failed-attempt counts for audit
  go to `audit_log` for post-mortem.

### 1.6 Cascade & retention

`ON DELETE CASCADE` on all FKs to `users(id)` — `DELETE /v1/me/data`
recursively removes profile + devices + projects + consent + reports +
api_tokens with one `DELETE FROM users WHERE id = $1`. Soft-delete
(`users.deleted_at`) is used only for login-blocking, not retention threshold.

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
- `meta String` — JSON object with safe extensions: `clipboard_sha256`,
  `clipboard_bytes`, `mouse_clicks`. No content (see [/privacy](/privacy) §1.1).
- TTL 18 months — auto-drop by partition. We do not use `DELETE WHERE` for
  retention (expensive); only for `DELETE /v1/me/data`.

### 2.2 Attribution — `attribution_events`

Derived from `events` via `attribution worker` (`backend/internal/attribution/`).
Categories here are more precise than raw:

```
category Enum8(
    'typed'        = 1,  -- appeared via keystroke stream
    'pasted_ai'    = 2,  -- paste from AI source (clipboard sha matched)
    'pasted_other' = 3,  -- paste without AI source
    'ai_inline'    = 4,  -- Copilot/Cursor inline accept
    'ai_agent'     = 5,  -- Claude Code / Cursor agent / Aider
    'refactor'     = 6,  -- structural change, not attributed
    'unknown'      = 7   -- worker could not classify
)
```

The worker runs in batches. Phase state (see comments in
[`backend/internal/attribution/worker.go`](../backend/internal/attribution/worker.go)):

**Phase A (now):**
1. reads fresh `events` with `category in (manual, ai)`;
2. maps each event to `attribution_events.category` by flat rules
   (`source`, `category`, `ai_channel` from raw event — no joins, no content);
3. writes batch to `attribution_events`, idempotent via dedup on `(ts, user_id, project_id)`.

**Phase B (roadmap):**
1. joins `clipboard_sha256` paste events with last-seen `clipboard_signal`
   (browser ext / desktop agent) — 30 sec window → `pasted_ai` vs `pasted_other`;
2. per-hunk attribution via diff snapshots in IDE extension;
3. precise inline/agent distinction without heuristics where direct API hooks exist.

### 2.3 Materialized aggregates

Cascading MVs (`backend/internal/migrate/sql/clickhouse/003_*`, `004_*`):

```
events (raw)  →  events_hourly_agg  →  events_daily_agg
              events_hourly_mv       events_daily_mv
```

- `events_hourly_agg` — `SummingMergeTree` on `(user_id, bucket_ts, category, file_lang)`.
  Hot queries: `AggregateByCategory`, `DailyTrend`, `LanguageBreakdown`, `Heatmap`.
- `events_daily_agg` — day bucket for long range (30/90 days).

**Routing** (`store/clickhouse.go`): `AggregateByCategory` / bulk → **daily** MV when
`since ≥ 30d`, else hourly. `DailyTrend`, `LanguageBreakdown`, `Heatmap` stay on hourly
(tz / day-of-week semantics).

**Reduction** (100 active users): raw ~10M rows/day → hourly ~290K (**35×**) → daily ~12K
(**840×** vs raw).

Materialized views roll up each INSERT in real time. Backfill on first migration apply;
on large prod DBs monitor `system.merges` — prefer chunked backfill by day/week.

**Benchmark** (`go run ./cmd/ch-bench` from `backend/`):

```sh
EOP_CH_DSN="clickhouse://default:@localhost:19000/eop_test" \
  go run ./cmd/ch-bench --users=10 --days=30 --events-per-user-per-day=10000
```

Local CH 24, 3M events (10×30×10K): `AggregateByCategory` p95 ~1 ms on daily MV vs ~8 ms
on hourly. Flags: `--skip-seed` if DB already seeded.

### 2.4 Redis read cache (`store/cached.go`)

Decorator over `EventStore` when Redis is reachable. TTLs: `AggregateByCategory` 10m,
`AggregateByCategoryBulk` 5m, `LanguageBreakdown` 10m, `DailyTrend` 5m, `Heatmap` 10m.
Keys: `eop:cache:{prefix}:{userID}:{params}`. Not cached: `Insert`, `ListRecent`,
`ActiveUserIDs`; `DeleteUserData` invalidates user keys. Redis down → miss + log, no fail.

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
                                                    │ (SummingMergeTree) │
                                                    └────────────────────┘

                                ┌─────────────────┐  batch SELECT
                                │ attribution-    │ ────────────► events
                                │ worker (Go)     │              (manual/ai)
                                └────────┬────────┘
                                         │ classify + dedup
                                         ▼
                                 ┌────────────────────┐
                                 │ attribution_events │
                                 │ (MergeTree)        │
                                 └────────────────────┘
```

## 4. Cross-cutting invariants

- **`user_id` is always UUIDv4**, never email or client-supplied. Backend
  overwrites `event.user_id` from JWT claims (`ingestapp/service.go`), even
  if the agent sent something else.
- **No content**. All tables store counters, hashes, or enum categories. `meta` —
  only safe keys (`clipboard_sha256`, …).
  More — [/privacy](/privacy) §1.1.
- **Idempotent migrations**. All `001..N` under advisory_lock + `IF NOT EXISTS`,
  rollbacks in `*.down.sql`. Hourly_agg backfill is safe to repeat.
- **TTL is the only retention mechanism** for analytics. We do not rely on
  background jobs to delete old events. GDPR erasure — explicit
  `ALTER TABLE ... DELETE WHERE user_id = ?`.

## 5. Extension

When adding a new field to Event:
1. Update `components.schemas.Event` in [`docs/api/openapi.yaml`](api/openapi.yaml).
2. Add field in `backend/internal/ingest/domain/event.go` and `backend/internal/store/event.go`.
3. Add column in `001_events.up.sql` via new migration `005_*.up.sql`
   (`ALTER TABLE events ADD COLUMN ... AFTER existing_col`).
4. Update `Insert()` in `clickhouse.go` (new positional arg in `batch.Append`).
5. If the field should aggregate — add to `events_hourly_agg` via
   next migration + new `MV`.
6. Agent-side: add field in `agent/src-tauri/src/core/event.rs`.

For safe extensions without content prefer `meta map<string,string>`
(already present; no migration required).
