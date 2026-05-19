# Backend agent rules

Эти правила обязательны для AI-агентов, работающих в `backend/`. Tech-lead-grade
expectations: build clean, lint clean, tests pass, no shortcuts.

## ⚠️ Pre-flight: ВСЕГДА запускай golangci-lint

**Перед любым `git commit` в `backend/`:**

```sh
cd backend && golangci-lint run ./...
```

Должно быть `0 issues`. Если linter ругается — **fix или explain**, не комитти.

CI gate (`.github/workflows/_backend.yml`) использует `golangci-lint v2.x`,
local config — `backend/.golangci.yml`. Если лень в локале — agent должен
прогнать сам перед commit и приложить output.

Помимо lint, минимальный pre-commit checklist:

```sh
go build ./...           # должен пройти без ошибок
go vet ./...             # zero output
go test ./...            # все unit-тесты pass (integration требуют PG/CH)
golangci-lint run ./...  # ★ блокер
```

Если падает любая стадия — diagnose root cause. **Не комитти broken state.**

## Documentation

- **Продуктовая и ops-документация для людей** — только в корневом [`docs/`](../docs/) (деплой, API, threat model, архитектура).
- **Не** дублировать в `backend/docs/` — один источник правды.
- Исключения: `backend/AGENTS.md` (правила для агентов), `README` пакета, промпты/runtime-конфиг в `internal/` (не «доки»).

## Architectural conventions

### Package layout

- `cmd/<name>/main.go` — bin entry-points (api, migrate, ingest, eop-hook,
  ch-bench, vapid-gen, changelog-gen). Каждый CLI tool — отдельный bin.
- `internal/<domain>/` — приватные packages, не importable снаружи modul'я.
  Используем `internal/`, **не `pkg/`** — у нас один module monorepo, нет
  external consumers (см. Q3 retro в ROADMAP).
- Domain split: `auth`, `teams`, `ingest`, `analytics`, `insights`, `webhooks`,
  `push`, `prcomment`, `anomaly`, `reports`, `store`, `cache`, `httperr`,
  `metrics`, `migrate`, `mailer`, `config`, `log`.

### Clean Architecture + DDD (incremental)

Цель: **bounded context** на домен (`internal/<domain>/`), внутри — слои с
зависимостями только внутрь (delivery → application → domain ← infrastructure).

| Слой DDD / CA | Пакет | Ответственность |
|---------------|--------|-----------------|
| **Domain** | `internal/<domain>/domain/` | Сущности, value objects, доменные ошибки, инварианты, **интерфейс репозитория** (порт persistence с точки зрения домена) |
| **Application** | `internal/<domain>/<slice>app/` или корневой `contentapp/` | Use cases / application services: оркестрация без Fiber/pgx в публичном API |
| **Ports (app)** | `ports.go` рядом с application | Интерфейсы: cache, audit, внешние ACL (super-admin), clock — всё, что не домен |
| **Infrastructure** | `store.go`, `cache.go`, `*_adapters.go` в корне домена | PG/Redis/mailer — реализации портов |
| **Delivery** | `handler.go` / `RegisterRoutes` | HTTP: parse → вызов application → `httperr` |

Правила:

- **Domain** не импортирует `fiber`, `pgx`, `zap`, другие bounded contexts.
- **Application** импортирует только `domain` + свои порты; не знает про SQL/Redis.
- **Delivery** тонкий: валидация transport (query/path/body), маппинг в DTO use case.
- Один **aggregate root** на срез там, где есть жизненный цикл и инварианты (напр. `content` → `domain.Block` по `slug+locale`).
- Ubiquitous language в именах: `Publish`, `SaveDraft`, `Revert`, не `Upsert` в HTTP handler.

Упрощённый эталон (полный DDD-слой): [`internal/content/`](internal/content/) — `domain/`, `contentapp/`, `handler.go`.

Классический тонкий срез (без отдельного `domain/`): [`internal/teams/emailtemplates/`](internal/teams/emailtemplates/) — порты + `Service` + wiring в [`emailtemplates_adapters.go`](internal/teams/emailtemplates_adapters.go).

Шаблон нового BC: [`internal/_template/`](internal/_template/). CI guard: `bash scripts/check-domain-imports.sh`.

#### Чеклист нового bounded context

1. `internal/<bc>/domain/` — entities, VOs, `errors.go`, repository/read ports при необходимости
2. `internal/<bc>/<bc>app/` (или `<slice>app/`) — `service.go`, `ports.go`, `service_test.go` с fake ports
3. `handler.go` — только Fiber + `httperr` + parse/marshal
4. `store.go` / `*_adapters.go` — pgx/CH/mailer ACL
5. `go test ./internal/<bc>/...` зелёный; integration при наличии
6. Строка в таблице миграции ниже: `domain + *app + handler`

**Naming:** один aggregate → `<bc>app` (`contentapp`); несколько срезов → `membersapp`, `invitesapp` под общим `teams/domain`. Не смешивать `teamapp` и `teamsapp` в одном BC.

**ACL между BC:** только application port в `ports.go` + реализация в `*_adapters.go` родителя; не импортировать чужой `handler.go`.

- Новые фичи — use case + тонкие handlers; legacy — **strangler** по одному срезу.
- Публичный audit API для IP: `audit.ClientIP(c *fiber.Ctx)` — общий helper
  для записи IP без дублирования X-Forwarded-For логики.

#### Статус миграции на use case-слои (по доменам)

Эталонный шаблон: [`internal/teams/emailtemplates/`](internal/teams/emailtemplates/) (порты, `Service`, unit-тесты с fake; wiring в [`emailtemplates_adapters.go`](internal/teams/emailtemplates_adapters.go); тонкие handlers).

| Домен | Срез / подпакет | Примечание |
|-------|-----------------|------------|
| content | `domain`, `contentapp`, `handler` | эталон (полный DDD) |
| analytics | `domain`, `analyticsapp`, `handler` | EventReadStore port + adapters |
| publicapi | `domain`, `publicapiapp`, `handler` | scoped public read API |
| ingest | `domain`, `ingestapp`, `handler` | PrepareBatch + PersistBatch |
| insights | `domain`, `insightsapp`, `rangeagg`, `handler` | Generate + fan-out |
| reports | `domain`, `reportsapp`, `periodapp`, `handler` | ReportGenerator port, cron → app |
| teams | `domain`, `teamsapp`, `membersapp`, `invitesapp`, `authapp`, `registrationapp`, `projectsapp`, `commitsapp`, `adminapp`, `emailtemplates`, … | thin `teams.go` / `members.go` / `commits.go`; `admin_handlers.go` |
| auth | `meapp`, `passwordresetapp`, `oauthapp`, `oauthflowapp`, `identitiesapp`, `passkeysapp`, `passwordapp`, `webauthnapp`, `sessionapp`, `apitokensapp` | `/v1/me/*` credentials + `/v1/auth/forgot|reset-password` in auth delivery |
| devices | `domain`, `devicesapp`, `pairingapp`, `handler` | |
| webhooks | `domain`, `webhooksapp` (CRUD + `Dispatcher`), `handler` | `webhooks.Service.Dispatch` → `webhooksapp.Dispatcher` |
| push | `domain`, `pushapp`, `handler` | |
| sso | `domain`, `ssoapp`, `handler` | |
| prcomment | `domain`, `prcommentapp`, `handler` | |
| anomaly | `anomalyapp` (scan + detect), `detector` | `cmd/api` wires `anomalyapp.Scanner` |
| attribution | `attributionapp`, `worker` | cmd/worker wiring |
| app | `internal/app` | composition root (`cmd/api` → `RegisterProductRoutes`) |

### Error handling

**Используй `internal/httperr` для HTTP error responses.** Pattern:

```go
// ❌ DON'T:
return c.Status(400).JSON(fiber.Map{"error": "invalid email"})

// ✅ DO:
return httperr.BadRequest(c, "invalid_email", "email malformed")
```

Helpers: `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`,
`Gone`, `TooLarge`, `TooManyRequests`, `Unavailable`, `BadGateway`, `Internal`.

Stable machine-readable `code` обязателен — frontend полагается на него для
i18n/handling. Для extra fields используй `httperr.Send(c, ProblemDetails{
Extensions: map[string]any{...}})`.

`Internal(c)` — generic detail "internal error" + auto request_id; **никогда
не leak'ает internal stack/error string наружу**. Caller должен `s.Logger.Error(...)`
сам перед.

### Concurrency

- **Bound concurrency** — никогда unbounded `g.Go()` в hot paths. Используй
  `errgroup.SetLimit(N)`:
  - DB-bound work: limit ≤ pool size (CH default pool = 10, set 8)
  - HTTP-bound: 4-8 (limits external service pressure)
- Fan-out pattern для independent work:
  ```go
  g, gctx := errgroup.WithContext(ctx)
  g.SetLimit(8)
  for _, item := range items {
      g.Go(func() error { /* using gctx */ return nil })
  }
  _ = g.Wait()
  ```
- Shared state (maps, counters) → `sync.Mutex` или `atomic.*`. См.
  `internal/anomaly/cron.go` `seenMu` pattern.
- Background goroutines (fire-and-forget) — отдельный `context.Background()`
  с timeout, не request ctx (request может cancellation'нуться).

### Database

- **pgx/v5** для Postgres. Connection pool через `pgxpool`.
- **clickhouse-go/v2** для CH. Pool через driver Options (см. `store/clickhouse.go`).
- `redis/v9` для Redis cache (optional, graceful degradation).
- Migrations: `golang-migrate/v4` с pgx5/clickhouse drivers (см. `internal/migrate`).
- **CH Cloud quirk:** `x-migrations-table-engine=MergeTree` (TinyLog запрещён
  на Shared databases) — auto-applied в `ensureCHMigrationsEngine`.
- **Multi-statement migrations:** `x-multi-statement=true` в DSN.
- `ON CONFLICT DO NOTHING` для idempotent inserts (commits ingestion).

### ClickHouse perf

- 3-tier cascading materialized views: `events` → `events_hourly_agg` (35×) →
  `events_daily_agg` (840× total). См. `docs/data-model.md` §2.3–2.4.
- Routing logic: long-range queries (≥30d) → daily MV; короткие — hourly.
- Redis cache layer (`internal/store/cached.go`) поверх для read-heavy
  aggregations с TTL 5-10 min.
- Bench tool: `go run ./cmd/ch-bench` для validation после изменений
  schema/queries.

### Testing

- **Unit tests** — без external deps, чистая Go logic. Run via `go test ./...`.
- **Integration tests** — `//go:build integration` build tag, требуют PG/CH.
  Test setup helpers: `setupTestDB(t)`, `createUser`, `loginAs`, `do`.
- **Coverage target Q1:** 40% (achieved 42.5% per ROADMAP).
- Run integration: `EOP_TEST_PG_DSN=... go test -tags=integration -p 1 ./...`
  (`-p 1` важно — teams + auth packages share test DB, parallel runs ломают).

### Logging

- **zap structured logger** через `internal/log`. Никаких `fmt.Println` /
  `log.Printf` в production code (только cmd/* CLI utilities).
- Pattern: `s.Logger.Error("operation failed", zap.Error(err), zap.String("user", uid))`.
- Не логируй PII: emails, passwords, full tokens. Token prefix (8 chars) OK.
- Request ID propagation — auto через `requestid` middleware; httperr автоматом
  включает в response.

### Security

- **No secrets в repo.** Config через env vars (`internal/config`).
- JWT validation в `auth.Middleware` — token_version revocation hook (демоут /
  password reset bumps tv → старые JWT инвалидируются).
- API tokens (eop_*): sha256 hash в DB, constant-time compare через
  `subtle.ConstantTimeCompare`.
- Rate limiting через `fiber/middleware/limiter` (см. `cmd/api/main.go`).
- CORS: explicit origins, no wildcard subdomain matching (Fiber CORS не
  поддерживает glob).
- HMAC webhooks: SHA-256, header `X-EoP-Signature: sha256=<hex>`.
- VAPID keys (Web Push) — ECDSA P-256, generated через `cmd/vapid-gen`,
  stored в env (NEVER commit private).

### Migrations

- **Up + down пары обязательны.** Production rollback требует.
- Naming: `NNN_description.up.sql` / `NNN_description.down.sql`.
- Migration applied auto на startup (`AutoMigrate=true` non-prod), prod
  через `cmd/migrate` runner.
- Backfill в той же migration где новая column/table — atomically.

## Code style

- **gofmt clean** (auto via lint).
- Imports grouped: stdlib → external → internal (`github.com/eye-of-providence/...`).
- Comments на русском или английском — consistent в одном файле.
- Doc comments на exported types/funcs — `// FuncName — описание...` (lowercase
  description after dash).
- Magic numbers → named constants с comment explaining choice.

## Pre-commit ritual

1. `go build ./...` — компилируется?
2. `go vet ./...` — статанализ?
3. `go test ./...` — unit-тесты pass?
4. `golangci-lint run ./...` — **0 issues?** ★
5. (если CH/PG migration) — `go run ./cmd/migrate up` локально + `down` rollback test
6. (если security-relevant) — `trivy image --severity HIGH,CRITICAL --ignorefile .trivyignore`
7. Commit с conventional-commit format (`feat(scope):`, `fix(scope):`, `perf(scope):`, etc.)

Если хоть одно failing — diagnose, не bypass. `--no-verify` запрещён.

## Anti-patterns (catch yourself)

- ❌ `fiber.Map{"error": ...}` — use httperr
- ❌ Unbounded `go func()` в loop — use errgroup with SetLimit
- ❌ `c.Status(500).JSON(fiber.Map{"error": err.Error()})` — leaks internal info
- ❌ `panic()` в request path — return error normally
- ❌ Logging через `log` package (stdlib) — use zap
- ❌ Tests без cleanup (TRUNCATE / t.Cleanup) — flaky tests
- ❌ SQL string concatenation с user input — use $1/$2 parameters
- ❌ `time.Sleep` в tests как sync mechanism — use channel / context.Done
- ❌ `import _ "side effect"` без commented justification
- ❌ Helper functions без doc-comment когда они exported
