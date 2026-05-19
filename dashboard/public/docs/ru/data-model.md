# Data model

Two stores: **Postgres** для transactional state (users, teams, devices,
projects, reports, audit log) и **ClickHouse** для analytics (events,
attribution, materialized aggregates). Источник правды wire-формата ingest —
[`docs/api/openapi.yaml`](api/openapi.yaml) (`components.schemas.Event`) и
Go-тип [`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go).
Транспорт: JSON `POST /v1/ingest` (не protobuf/gRPC).

Все SQL-миграции под версионным контролем в `backend/internal/migrate/sql/`
(`postgres/NNN_*.up.sql` и `clickhouse/NNN_*.up.sql`). API применяет их
автоматически на старте, если `EOP_AUTO_MIGRATE=true`.

## 1. Postgres

### 1.1 Identity & ownership

| Таблица | Назначение | Notable |
|---|---|---|
| `users` | Аккаунты + soft-delete (`deleted_at`) | `email UNIQUE`, OAuth `github_login`, `team_id` FK, `token_version` для revocation |
| `teams` | Группы пользователей + plan (`free`/`pro`/...) | `settings_json JSONB` для flexible per-team config |
| `team_members` | Многие-ко-многим users↔teams + role | См. `003_multi_team_projects.up.sql` |
| `webauthn_credentials` | FIDO2/passkey credentials | См. `020_webauthn_credentials.up.sql` |

### 1.2 Devices & projects

| Таблица | Поле | Зачем |
|---|---|---|
| `devices` | `(user_id, fingerprint) UNIQUE` | Привязка десктоп-агента; pair-flow выдаёт row здесь |
| `projects` | `root_path_hash` (sha256), не сам путь | Per-project attribution без раскрытия FS |

### 1.3 Consent & API tokens

- `consent` — granted/revoked timestamps по scope (`telemetry`, `ai_reports`, …).
- `api_tokens` — bearer-токены для CI/scripts; `hashed_token TEXT NOT NULL UNIQUE`
  (bcrypt). Plaintext возвращается ровно один раз на создании.

### 1.4 Reports & audit

- `reports` — AI-сгенерированные markdown-отчёты, `period` ∈ `weekly_2026_W18`/`monthly_2026_05`,
  `prompt_version` для отслеживания AB-вариаций промпта.
- `audit_log` (миграция `016_audit_log.up.sql`) — security-sensitive
  events: `login_failed`, `password_reset`, `token_created/revoked`,
  `user_deleted`. Retention 24 мес.

### 1.5 Webhooks & rate-limit

- `webhooks` (`011_webhooks.up.sql`) — outbound HTTP-уведомления (Slack/
  Discord) с retry-state и HMAC-секретом.
- Rate-limit state — в Redis, не в Postgres. Audit failed-attempt counts
  идут в `audit_log` для post-mortem.

### 1.6 Cascade & retention

`ON DELETE CASCADE` на всех FK к `users(id)` — `DELETE /v1/me/data`
рекурсивно сносит profile + devices + projects + consent + reports +
api_tokens одним `DELETE FROM users WHERE id = $1`. Soft-delete
(`users.deleted_at`) используется только для login-blocking, не для
порога retention.

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
- `meta String` — JSON-объект с безопасными расширениями: `clipboard_sha256`,
  `clipboard_bytes`, `mouse_clicks`. Контента нет (см. `docs/privacy.md` §1.1).
- TTL 18 мес — auto-drop по partition. Не используем `DELETE WHERE` для
  retention (дорого); только для `DELETE /v1/me/data`.

### 2.2 Attribution — `attribution_events`

Derived из `events` через `attribution worker` (`backend/internal/attribution/`).
Категории здесь точнее, чем в raw:

```
category Enum8(
    'typed'        = 1,  -- появилось через keystroke stream
    'pasted_ai'    = 2,  -- paste из AI-источника (clipboard sha совпал)
    'pasted_other' = 3,  -- paste без AI-источника
    'ai_inline'    = 4,  -- Copilot/Cursor inline accept
    'ai_agent'     = 5,  -- Claude Code / Cursor agent / Aider
    'refactor'     = 6,  -- структурное изменение, не attribut'нуто
    'unknown'      = 7   -- worker не смог классифицировать
)
```

Worker запускается батчами. Состояние по фазам (см. комментарии в
[`backend/internal/attribution/worker.go`](../backend/internal/attribution/worker.go)):

**Phase A (сейчас):**
1. читает свежие `events` с `category in (manual, ai)`;
2. маппит каждое событие на `attribution_events.category` по плоским правилам
   (`source`, `category`, `ai_channel` из raw event — без join'ов и без контента);
3. пишет batch в `attribution_events`, idempotent через dedup по `(ts, user_id, project_id)`.

**Phase B (roadmap):**
1. джойнит `clipboard_sha256` paste-событий с last-seen `clipboard_signal`
   (browser ext / desktop agent) — окно 30 сек → `pasted_ai` vs `pasted_other`;
2. per-hunk attribution через diff snapshots в IDE extension;
3. точное различение inline/agent без эвристик там, где есть direct API hooks.

### 2.3 Materialized aggregates

Cascading MVs (`backend/internal/migrate/sql/clickhouse/003_*`, `004_*`):

```
events (raw)  →  events_hourly_agg  →  events_daily_agg
              events_hourly_mv       events_daily_mv
```

- `events_hourly_agg` — `SummingMergeTree` по `(user_id, bucket_ts, category, file_lang)`.
  Hot queries: `AggregateByCategory`, `DailyTrend`, `LanguageBreakdown`, `Heatmap`.
- `events_daily_agg` — day-bucket для long-range (30/90 days).

**Routing** (`store/clickhouse.go`): `AggregateByCategory` / bulk → **daily** MV when
`since ≥ 30d`, else hourly. `DailyTrend`, `LanguageBreakdown`, `Heatmap` stay on hourly
(tz / day-of-week semantics).

**Reduction** (100 active users): raw ~10M rows/day → hourly ~290K (**35×**) → daily ~12K
(**840×** vs raw).

Materialized views roll up each INSERT in real-time. Backfill on first migration apply;
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

- **`user_id` всегда UUIDv4**, никогда не email и не client-supplied. Backend
  переписывает `event.user_id` из JWT-claims (`ingestapp/service.go`), даже
  если агент прислал что-то другое.
- **Никакого контента**. Все таблицы хранят либо счётчики, либо хеши, либо
  enum-категории. `meta` — только безопасные ключи (`clipboard_sha256`, …).
  Подробнее — [`docs/privacy.md` §1.1](privacy.md#11-никогда-не-покидает-машину-пользователя).
- **idempotent miграции**. Все `001..N` под advisory_lock + `IF NOT EXISTS`,
  rollback'и в `*.down.sql`. Backfill в hourly_agg повторно-безопасен.
- **TTL — единственный механизм retention** для analytics. Не полагаемся на
  background-jobs для удаления старых событий. GDPR-erasure — explicit
  `ALTER TABLE ... DELETE WHERE user_id = ?`.

## 5. Расширение

При добавлении нового поля в Event:
1. Обнови `components.schemas.Event` в [`docs/api/openapi.yaml`](api/openapi.yaml).
2. Добавь поле в `backend/internal/ingest/domain/event.go` и `backend/internal/store/event.go`.
3. Добавь column в `001_events.up.sql` через новую миграцию `005_*.up.sql`
   (`ALTER TABLE events ADD COLUMN ... AFTER existing_col`).
4. Обнови `Insert()` в `clickhouse.go` (новый поз. arg в `batch.Append`).
5. Если поле должно агрегироваться — добавь в `events_hourly_agg` через
   следующую миграцию + новый `MV`.
6. Agent-side: добавь поле в `agent/src-tauri/src/core/event.rs`.

Для безопасных расширений без контента предпочитай `meta map<string,string>`
(уже есть; не требует миграции).
