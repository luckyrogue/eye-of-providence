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
  `events_daily_agg` (840× total). См. `docs/clickhouse-perf.md`.
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
