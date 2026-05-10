# E2E suite (Playwright)

Full-stack end-to-end tests: реальный browser → dashboard (Vite preview) →
backend (Go) → PG/CH/Redis (docker compose).

## Quick start (local)

```sh
# 1) Поднять инфру (PG/CH/Redis):
pnpm infra:up

# 2) Поставить browsers (один раз):
pnpm -F @eop/e2e exec playwright install --with-deps chromium

# 3) Build dashboard с правильным backend URL:
cd dashboard && VITE_BACKEND_URL=http://localhost:8080 pnpm build && cd ..

# 4) Запустить tests (playwright сам поднимет backend + preview):
pnpm -F @eop/e2e test
```

Опции:

```sh
pnpm -F @eop/e2e test:ui       # interactive UI mode (debugger)
pnpm -F @eop/e2e test:headed   # headed browser (visible)
pnpm -F @eop/e2e report        # открыть HTML report после run'а
```

## Что покрывается

| Spec | What |
|---|---|
| `health.spec.ts` | Backend `/healthz`, dashboard renders |
| `auth.spec.ts` | Register/login/forgot-password, typed errors |
| `onboarding.spec.ts` | Status / dismiss / locale |
| `teams.spec.ts` | CRUD + invite link flow + owner_limit |
| `ingest.spec.ts` | Event accept/reject, batch limit, ListRecent |
| `analytics.spec.ts` | Categories/languages/trend/heatmap |
| `insights.spec.ts` | Fan-out insights endpoint |
| `api-tokens.spec.ts` | Create/use/scope/revoke |
| `webhooks.spec.ts` | CRUD |
| `publicapi.spec.ts` | Read-only public endpoints + scope checks |
| `admin.spec.ts` | super_admin guards + promote via psql |
| `dashboard-ui.spec.ts` | Browser smoke: login → dashboard → settings |

## NE покрывается

- **PR comment** (`internal/prcomment`) — требует GitHub/GitLab API mock'а
- **Reports** (`internal/reports`) — требует Gemini API key
- **Mailer/email-send** — Resend SMTP не моки'м
- **PR push delivery** (webhooks outbound) — требует receiver

Эти domain'ы покрыты integration-тестами в `backend/internal/*/...`.

## CI

Workflow `.github/workflows/e2e.yml` — на каждый PR + main push:

1. `docker compose up -d` (PG/CH/Redis)
2. Setup Go 1.26 + pnpm + Node 20
3. `pnpm install --frozen-lockfile`
4. Build dashboard (`VITE_BACKEND_URL=http://localhost:8080`)
5. Start backend (Go `cmd/api`, `EOP_AUTO_MIGRATE=true`)
6. `playwright install chromium`
7. `pnpm -F @eop/e2e test`
8. Upload `playwright-report/` как artifact на fail

## Anti-flake patterns

- **Unique emails per test** — `uniqueEmail("prefix")` генерит уникальный
  `e2e-{prefix}-{ts}-{rand}@local.test`. Каждый test изолирован.
- **CH propagation polling** — `for (let attempt = 0; attempt < 10; ...)`
  + 500ms sleep'ы для запросов после ingest (events async flush).
- **Single worker** (`workers: 1`) — shared DB state, parallel прогон ломает
  Postgres `users` row counts (e.g. first-user bootstrap).
- **`reuseExistingServer: true` локально** — не пересоздаём процессы
  между прогонами.

## Adding a test

```ts
import { test, expect } from "../fixtures/index.js";

test("my new test", async ({ session, api }) => {
  // `session` — авто-создан JWT юзер
  // `api` — ApiClient с auto-token'ом
  const r = await api.fetch("/v1/my/endpoint");
  expect(r).toBeDefined();
});
```

Если test'у нужен второй юзер — `apiRegister(uniqueEmail("foo"), "pass")` +
`createApiClient(otherToken)`.

## Debug failing test

```sh
# Trace viewer:
pnpm -F @eop/e2e exec playwright show-trace test-results/.../trace.zip

# Запустить один test:
pnpm -F @eop/e2e test -g "create team"

# С debug log'ом:
PWDEBUG=1 pnpm -F @eop/e2e test
```
