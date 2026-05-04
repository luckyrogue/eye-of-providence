# Implementation Plan — Eye of Providence

Детализированный план: что и когда делаем, с конкретными задачами, артефактами и критериями готовности. Опирается на архитектуру из `README.md`.

---

## Принципы исполнения

1. **Vertical slices, не layers** — каждую неделю одна end-to-end фича работает (event течёт от агента до дашборда), а не "сначала весь backend, потом весь frontend".
2. **Privacy-first с первого коммита** — никогда не пишем код, который собирает контент; redaction до того, как payload покидает память агента.
3. **Skeleton-first, polish later** — на MVP важнее, что все слои соединены, а не красота каждого.
4. **Один общий Protobuf-контракт** для событий (rust agent → go backend → ts ui) — чтобы schema-drift ловился compile-time.

---

## Phase 0 — Foundation (день 1–3)

Цель: монорепо, CI, общие контракты, заглушки сервисов.

### Tasks
- [ ] `git init` + monorepo structure (см. README §12).
- [ ] `proto/` — первая версия Protobuf-схемы:
  - `Event` (ts, user_id, device_id, app_bundle, category, source, ai_provider, duration_ms, …).
  - `IngestRequest` (batch of events).
  - `Report` (id, period, body_md, …).
- [x] Code generation pipeline: `buf` config готов в `proto/buf.gen.yaml`. Реальная codegen — отложена, типы пишем вручную (быстрее на skeleton-стадии).
- [x] `ui/` — package с shadcn/ui Button + Card, Tailwind preset. Используется в agent/dashboard/browser-extension.
- [x] `backend/` — Go module + Fiber + zap, `/healthz`, конфиг через env (FromEnv).
- [x] `agent/` — Tauri 2 + React UI, `cargo tauri` config готов (cargo на этой машине нет, проверка в CI).
- [x] `infra/docker-compose.yml` — Postgres, ClickHouse, Redis (verified working).
- [x] GitHub Actions: `.github/workflows/ci.yml` (Go vet/build, frontend build, Rust check на macos-latest).

### Готово, когда:
- [x] `docker compose up` поднимает Postgres + ClickHouse + Redis локально.
- [x] `cargo tauri dev` — конфиг готов; для сборки нужен Rust toolchain.
- [x] `go run ./cmd/api` отвечает на `/healthz`.
- [x] CI workflow зарегистрирован.

---

## Phase 1 — Skeleton end-to-end (неделя 1–2)

Цель: один event ходит от macOS агента до дашборда.

### Agent (macOS only пока)
- [x] Rust core: trait `PlatformWatcher`, watcher loop, в Phase 1 без keystrokes (Phase 1.5).
- [x] `platform/macos/`: `NSWorkspace.frontmostApplication` (objc2), `CGEventSourceSecondsSinceLastEventType` для idle (FFI).
- [x] Локальный SQLite буфер событий (`rusqlite` WAL mode), lease-based queue.
- [x] HTTP client → backend `/v1/ingest` (`reqwest`, bearer token).
- [ ] Onboarding flow: Accessibility permission. **Phase 1.5** — нужно для keystrokes.
- [ ] Tauri tray icon. **Phase 1.5**.

### Backend
- [x] `backend/internal/auth/`: dev-token (mock JWT), GitHub OAuth callback.
- [x] `backend/internal/ingest/`: Fiber `/v1/ingest`, JWT auth, privacy-валидация event'ов.
- [x] `backend/internal/analytics/`: `/v1/events/recent`, `/v1/summary/categories`.
- [x] `backend/cmd/api/`: единый dev-сервер (auth + ingest + analytics).
- [x] Postgres миграции: `users`, `devices`, `consent`, `reports` (готовы из Phase 0).
- [x] In-memory store как dev fallback. ClickHouse adapter — Phase 2.
- [x] zap-логгер, CORS для dashboard.

### Dashboard
- [x] Vite + React SPA, dev-token login flow.
- [x] Экран: AI ratio + последние события (table) + send-demo кнопка.
- [x] API client (TS) с типами событий.

### Готово, когда:
- [x] **Smoke-test PASSED**: dev-token → POST /v1/ingest → GET /v1/events/recent → GET /v1/summary/categories.
- [x] Dashboard собирается, поднимается на `:5174`, фетчит события из backend.
- [ ] Real macOS agent пишет реальное focus-event в backend. **Требует cargo/rust toolchain — пока не верифицировано на этой машине; код готов**.

---

## Phase 2 — Browser extension + AI detection (неделя 2–3)

### Browser extension
- [x] MV3 manifest, Vite + CRX, TypeScript + React, Tailwind + shadcn popup.
- [x] Background service worker: `chrome.tabs.onActivated`, `chrome.windows.onFocusChanged`, `chrome.idle`, alarms-based flush каждые 30с.
- [x] Content scripts для AI-сайтов (chatgpt, claude, gemini, copilot, perplexity).
- [x] Copy event detection: heuristic-селекторы `data-message-author-role="assistant"` и т.д., шлём только `host` + `size` (не контент).
- [x] AI domain → provider/channel mapping.
- [x] API client → backend через JWT в `chrome.storage.local`.
- [x] Popup на shadcn: dev login + flush-now + logout.
- [ ] Native messaging host для связи с desktop agent. **V1+** — для Phase 2 работает напрямую с backend.

### Готово, когда:
- [x] Extension собирается (`pnpm -F @eop/browser-extension build`), валидный MV3 dist/.
- [x] Privacy gate: размеры copy событий, не контент.
- [ ] Реальный chrome загрузил unpacked extension и события появляются в dashboard. **Требует ручную загрузку в browser**.

---

## Phase 3 — Attribution v1 + IDE plugin (неделя 3–4)

### VS Code extension
- [x] TS extension: `onDidChangeTextDocument`, `onDidChangeActiveTextEditor`, `onDidSaveTextDocument`.
- [x] Per-edit классификация: малые insert (< 80 chars) → `manual`, большие insert → `ai/inline` (proxy для Copilot/Cursor accept), large replace → `refactor`.
- [x] Per-language buckets с агрегацией chars/lines/duration.
- [x] Periodic flush (default 30с) + `eop.flush` команда.
- [x] `eop.devLogin` команда: GET dev-token и сохраняет в settings.
- [x] Backend URL и paste threshold конфигурируются через VS Code settings.

### Claude Code hooks
- [x] `cli-hooks/eop-claude-hook.sh` — приватная обёртка: НЕ читает stdin (не пересылает контент), шлёт только факт срабатывания (category=ai, source=cli, provider=anthropic, channel=agent).
- [x] `cli-hooks/claude-code-install.sh` — jq-based merge в `~/.claude/settings.json` для Stop / PostToolUse / UserPromptSubmit, с backup.
- [x] Documented env: `EOP_TOKEN`, `EOP_URL`.

### Attribution worker (Go)
- [x] **Перенесли в client-side**: VS Code extension классифицирует hunks по эвристикам сразу. Backend worker превратится в server-side enrichment в Phase 4 (когда подключим real ClickHouse и attribution_events table).

### Готово, когда:
- [x] **Smoke-test PASSED**: IDE event (manual + ai/inline) + CLI hook event (claude-code stop) → backend → recent endpoint.
- [x] VS Code extension компилируется (`tsc -p .` → `dist/extension.js`).
- [ ] Дашборд показывает AI ratio с учётом chars (а не только ms). **Phase 4** — нужны улучшенные аналитические запросы.

---

## Phase 4 — Windows parity (неделя 4–5)

- [ ] `platform/windows/`: `SetWinEventHook`, `GetForegroundWindow` + `QueryFullProcessImageName`, `GetLastInputInfo`.
- [ ] Keystroke counts через `SetWindowsHookEx(WH_KEYBOARD_LL)` (low-level, низкий риск).
- [ ] Clipboard через `GetClipboardSequenceNumber` + `OpenClipboard` (только хеш + размер).
- [ ] MSI/EXE builder в Tauri config, code signing pipeline (отложить notarization).
- [ ] Autostart через registry `Run`.
- [ ] Контракт-тесты: один и тот же event-payload генерится на обеих ОС для эквивалентного сценария.

### Готово, когда:
- На Windows-машине агент работает с тем же поведением, что на macOS. Дашборд показывает обе device-platform.

---

## Phase 5 — AI reports (Gemini) + polish (неделя 5–6)

### Reports service
- [x] `backend/internal/reports/`: aggregate, gemini client, store, handler.
- [x] `POST /v1/reports/generate?period=weekly|monthly|daily` — собирает агрегаты из EventStore → Gemini.
- [x] `GET /v1/reports/` — список, `GET /v1/reports/:id` — конкретный.
- [x] **Gemini REST API** (без SDK): `gemini-2.5-flash` через `generativelanguage.googleapis.com/v1beta`, system prompt embeded из `prompts/system.md`.
- [x] **Mock mode** когда `EOP_GEMINI_API_KEY` пуст: детерминистичный отчёт с реальными агрегатами — для UI dev без API key.
- [x] In-memory ReportStore + Postgres persistence (Phase 6).
- [x] Cron weekly autotrigger (Phase 6.5: `internal/reports/cron.go`, `EOP_REPORTS_CRON_SEC`).

### Privacy guarantees
- [x] `BuildContext` собирает только числа: durations (sec), chars per provider/language/category. **Никаких** URL, file paths, prompt text, response content.
- [x] `buildUserPrompt` шлёт в Gemini stripped JSON (только агрегаты + period dates).
- [x] System prompt запрещает выдумывать числа и угадывать контент.

### Dashboard polish
- [x] 3-card grid: AI ratio %, Active min, Reports count.
- [x] AI report card: Generate weekly / Monthly buttons, отчёт-switcher (chip-список periods), markdown rendering без зависимостей.
- [x] Recent events table с send-demo и refresh.
- [x] Heatmap часы×дни. **Сделано в Phase 6.5** (`Heatmap.tsx` + `/v1/heatmap`).
- [ ] AI ratio trend chart 7/30/90д. **V1+** — нужны time-series rollups.

### Готово, когда:
- [x] **Smoke-test PASSED**: 6 событий → POST /v1/reports/generate?period=weekly → markdown с TL;DR / Ключевые цифры / AI по провайдерам / Top языков / Рекомендация.
- [x] Dashboard рендерит markdown через минимальный custom Markdown компонент.
- [x] Mock-mode позволяет демонстрировать flow без Gemini API key.

---

## Cross-cutting (всё время)

- [ ] **Telemetry of telemetry**: метрики работы агента (event drop rate, latency до backend, ошибки) — в собственный бакет, агрегаты только. **V1**.
- [x] **Documentation**: `docs/privacy.md`, `docs/attribution.md`, `docs/data-model.md`, `docs/self-hosting.md` — есть, self-hosting расширен в Phase 7.
- [ ] **Threat model**: один проход через STRIDE — **отложено до V1**.
- [ ] **Cost dashboard**: Gemini token usage, ClickHouse storage — **V1**.

---

## Phase 6 — Real persistence (ClickHouse + Postgres) + heatmap

### EventStore: ClickHouse
- [x] `internal/store/clickhouse.go`: `OpenClickHouse(dsn)`, batch INSERT, `ListRecent`, `AggregateByCategory`, `Heatmap`.
- [x] DSN `clickhouse://user:pass@host:port/db` parsing.
- [x] UUID-aware: user/device/session/project columns — UUID, остальное LowCardinality(String).

### ReportStore: Postgres
- [x] `internal/reports/postgres_store.go`: `OpenPostgres`, `NewPostgresStore(pool)`, `Save/ListForUser/Get`.
- [x] FK constraint `reports.user_id → users.id` enforced.
- [x] Auto-upsert user в `users` через `auth/users_pg.go` для dev-token и github-callback.

### Analytics
- [x] `GET /v1/heatmap?days=N` — DOW × hour × category, агрегаты часов в ms.
- [x] ClickHouse: `toDayOfWeek(ts) → 1..7 → 0..6`, `toHour(ts)`.
- [x] In-memory equivalent: тот же контракт, для dev без Docker.

### Fallback strategy
- [x] Backend поднимается без Docker (in-memory сторы), переключается автоматически если `EOP_POSTGRES_DSN` / `EOP_CLICKHOUSE_DSN` указывают на доступную базу.
- [x] Логи показывают какой store выбран.

### Готово, когда:
- [x] **E2E PASSED**: user upsert в Postgres → 2 события через ingest → ClickHouse 2 строки → POST /reports/generate → Postgres reports row с body_md (456 chars).
- [x] `GET /v1/heatmap?days=7` возвращает корректные cells.

---

## Phase 6.5 — heatmap UI + per-language card + cron

### Backend
- [x] `GET /v1/summary/languages?days=N` — `[{lang, category, chars, ms}]` через ClickHouse `GROUP BY file_lang, category` (для in-memory тот же контракт).
- [x] `EventStore.ActiveUserIDs(since)` — `SELECT DISTINCT user_id` для cron.
- [x] `internal/reports/cron.go` — фоновый `Cron.Run(ctx)`: тикает `EOP_REPORTS_CRON_SEC` секунд, для каждого активного user проверяет если weekly отчёт за текущую ISO week уже есть, иначе генерирует.
- [x] Конфиг `EOP_REPORTS_CRON_SEC` (0 = выключено по умолчанию).

### Dashboard
- [x] `Heatmap.tsx` — 7×24 grid, color intensity = total ms, hover показывает breakdown per category.
- [x] `Languages.tsx` — top-N languages, bar manual/ai/refactor/other, % AI.
- [x] App.tsx: новые карточки "Activity heatmap" (lg:grid-cols-2) и "By language", `fetchLanguages(30)` + `fetchHeatmap(30)` в refresh.

### Готово, когда:
- [x] **E2E PASSED**: 3 события (TS manual + TS AI inline + Go manual) → `/v1/summary/languages` возвращает 3 cells (typescript manual+ai, go manual), heatmap → 2 cells (dow=1 hour=16 ai/manual).
- [x] Cron logged: `cron: generated user=… period=weekly_2026_W19` после старта.
- [x] Dashboard собирается, preview на :5174.

---

## Phase 7 — production deployability + delete-my-data + settings UI

### Backend
- [x] `backend/Dockerfile` — multi-stage build, distroless final image (~27 MB), nonroot user.
- [x] `infra/docker-compose.full.yml` — postgres + clickhouse + redis + api в одном compose, env-driven secrets.
- [x] `internal/store/delete.go` — `UserDeleter` capability: ClickHouse `ALTER TABLE events DELETE WHERE user_id = ?`, in-memory equivalent.
- [x] `internal/auth/me.go` — `GET /v1/me/`, `DELETE /v1/me/data` (тx-обёртка стирающая reports/api_tokens/consent/projects/devices/users).
- [x] `docs/self-hosting.md` — пошаговый quickstart, env-таблица, production checklist.

### Dashboard
- [x] `Settings.tsx` — Profile / Privacy / Danger zone (Delete all my data с подтверждением).
- [x] App.tsx: tab-switcher Dashboard | Settings, Logout button.

### Готово, когда:
- [x] **E2E PASSED**: `DELETE /v1/me/data` → 1→0 user в pg, 1→0 reports, 1→0 events в ClickHouse.
- [x] `docker build -t eop-api .` собирает образ 27 MB.
- [x] Dashboard собирается с tab-switcher.

---

## После MVP — что не делаем сейчас

Сознательно вне scope MVP:
- JetBrains plugin.
- Mobile app.
- Team mode (admin dashboard) — только основа, без UI.
- Self-host docker-compose как продукт (только для dev).
- Local AI reports через Ollama.
- SOC2 / DPA процессы.

Эти пункты — V1+ из README §9.

---

## Открытые решения (нужны до Phase 1)

1. **Хостинг бэкенда на старте**: Fly.io (быстрее) vs k8s в GCP (сразу production-ready). Рекомендация: Fly.io до 100 пользователей.
2. **Code signing certs**: нужны Apple Developer ID ($99/год) и Windows EV cert ($200–500/год). Купить в Phase 0.
3. **Gemini API доступ**: создать GCP проект, получить ключ, оценить квоту. Сделать в Phase 0, чтобы Phase 5 не упёрся.
4. **Бренд / домен / лого**: блокирует OAuth setup (нужны redirect URIs). Решить до Phase 1.

---

## First commit — что сделать прямо сейчас

1. `git init` в `eye-of-providence/`.
2. Создать структуру папок из README §12 (пустые с `.gitkeep`).
3. `proto/event.proto` — первая схема.
4. `infra/docker-compose.yml` — postgres, clickhouse, redis.
5. `backend/go.mod` + Fiber `/healthz`.
6. `agent/` — `cargo tauri init`.
7. `ui/` — `pnpm create vite` + shadcn init.
8. CI workflow — `lint.yml`.

Начинать?
