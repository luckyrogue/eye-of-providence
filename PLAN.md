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
- [ ] MV3 manifest, Vite + CRX, TypeScript + React.
- [ ] Background service worker: `chrome.tabs`, `chrome.windows.onFocusChanged`, visibility tracking.
- [ ] Content scripts для AI-сайтов (chatgpt.com, claude.ai, gemini.google.com, copilot.microsoft.com, perplexity.ai).
- [ ] Copy event detection: listener на сообщения ассистента, sha256 содержимого + размер.
- [ ] Native messaging host (через локальный `127.0.0.1` + одноразовый pairing-код) для связи с desktop agent.
- [ ] Popup на shadcn (статус, pause, last 24h summary).

### Готово, когда:
- Время на ChatGPT.com засчитывается отдельной категорией `ai/browser-chat`.
- Copy из Claude → событие `clipboard_ai_source` ушло в backend.

---

## Phase 3 — Attribution v1 + IDE plugin (неделя 3–4)

### VS Code extension
- [ ] TS extension: `onDidChangeActiveTextEditor`, `onDidChangeTextDocument`, `onDidSaveTextDocument`.
- [ ] Snapshot + diff на каждый save; hunks с timestamps.
- [ ] Copilot inline accept hook (если API доступен) или эвристика (большой одномоментный insert).
- [ ] Отправка attribution-events в desktop agent (тот же `127.0.0.1` канал).

### Attribution worker (Go)
- [ ] `backend/cmd/worker/` — фоновый job, читает raw events из CH, классифицирует hunks по правилам из README §4.2.
- [ ] Категории: `typed | pasted-AI | pasted-other | AI-inline | AI-agent | unknown`.
- [ ] Запись в `attribution_events` table.

### Claude Code hooks
- [ ] Скрипт-генератор: добавляет в `~/.claude/settings.json` хуки `Stop`, `PostToolUse`, `UserPromptSubmit` → POST в local agent.
- [ ] Документация: how to enable.

### Готово, когда:
- На дашборде видно график `manual vs AI ratio` с разбивкой по провайдерам.

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
- [ ] `backend/cmd/reports/` — Fiber endpoint `/v1/reports/generate` + cron (weekly/monthly).
- [ ] Aggregator: ClickHouse query → структурированный numeric context (JSON, без контента).
- [ ] Gemini integration через `google.golang.org/genai`, model `gemini-2.5-flash`, context caching системного промпта.
- [ ] Prompt templates в `backend/internal/reports/prompts/`:
  - `system.md` — роль, формат markdown, запрет на выдумывание.
  - `weekly_personal.md` — few-shot с примерами хороших отчётов.
  - `monthly_personal.md`.
- [ ] Persist в Postgres `reports(id, user_id, period, model, body_md, generated_at, prompt_version)`.

### Dashboard polish
- [ ] Heatmap (часы × дни, color = active time).
- [ ] AI ratio trend chart (7/30/90д).
- [ ] Per-language / per-project breakdown.
- [ ] Кнопка "Generate AI report" + просмотр прошлых отчётов.
- [ ] Settings: opt-in для AI reports с экраном "что именно отправляется в Google".

### Privacy gates
- [ ] Audit на ingest: payload без content fields (compile-time через тип `Event` без content полей; runtime тест).
- [ ] Аудит на reports: numeric context — золотой тест, что в Gemini шлются только числа и метки.
- [ ] "Delete all my data" endpoint, проверка что чистит CH/PG/S3.

### Готово, когда:
- Пользователь тыкает "Generate weekly report" → через 10 сек появляется markdown с инсайтами.
- Privacy-аудит-тест зелёный.
- Onboarding flow от signup до первого отчёта работает без ручных команд.

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
