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
- [ ] Code generation pipeline: `buf` или `protoc` → Rust (`prost`), Go (`protoc-gen-go`), TS (`ts-proto`).
- [ ] `ui/` — package boilerplate: Vite + React + Tailwind + shadcn/ui (`npx shadcn@latest init`), 3 базовых компонента (Button, Card, Chart-wrapper).
- [ ] `backend/` — Go module, Fiber + zap скелет, `/healthz`, конфиг через env.
- [ ] `agent/` — `cargo tauri init`, React UI с одним экраном "Hello".
- [ ] `infra/docker-compose.yml` — Postgres, ClickHouse, Redis для локальной разработки.
- [ ] GitHub Actions: lint + test + build для каждого пакета.

### Готово, когда:
- `docker compose up` поднимает всё локально.
- `cargo tauri dev` показывает окно с React UI.
- `go run ./cmd/ingest` отвечает на `/healthz`.
- CI зелёный.

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
- [x] In-memory ReportStore. **Phase 6** — Postgres persistence.
- [ ] Cron weekly/monthly autotrigger. **Phase 6** — после Postgres.

### Privacy guarantees
- [x] `BuildContext` собирает только числа: durations (sec), chars per provider/language/category. **Никаких** URL, file paths, prompt text, response content.
- [x] `buildUserPrompt` шлёт в Gemini stripped JSON (только агрегаты + period dates).
- [x] System prompt запрещает выдумывать числа и угадывать контент.

### Dashboard polish
- [x] 3-card grid: AI ratio %, Active min, Reports count.
- [x] AI report card: Generate weekly / Monthly buttons, отчёт-switcher (chip-список periods), markdown rendering без зависимостей.
- [x] Recent events table с send-demo и refresh.
- [ ] Heatmap часы×дни. **Phase 6** — нужны server-side aggregations.
- [ ] AI ratio trend chart 7/30/90д. **Phase 6**.

### Готово, когда:
- [x] **Smoke-test PASSED**: 6 событий → POST /v1/reports/generate?period=weekly → markdown с TL;DR / Ключевые цифры / AI по провайдерам / Top языков / Рекомендация.
- [x] Dashboard рендерит markdown через минимальный custom Markdown компонент.
- [x] Mock-mode позволяет демонстрировать flow без Gemini API key.

---

## Cross-cutting (всё время)

- [ ] **Telemetry of telemetry**: метрики работы агента (event drop rate, latency до backend, ошибки) — в собственный бакет, агрегаты только.
- [ ] **Documentation**: `docs/privacy.md`, `docs/attribution.md`, `docs/data-model.md`, `docs/self-hosting.md`.
- [ ] **Threat model**: один проход через STRIDE до конца Phase 1, ещё один до Phase 5.
- [ ] **Cost dashboard**: Gemini token usage, ClickHouse storage — чтобы не словить сюрприз.

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
