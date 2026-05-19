# Eye of Providence

> **Status:** v0.1 · публичная альфа · macOS + Windows + Linux ([changelog](CHANGELOG.md))
> Current release: [`v0.1.0-alpha.2`](https://github.com/luckyrogue/eye-of-providence/releases/tag/v0.1.0-alpha.2)

Система отслеживания использования AI при разработке: измеряет, сколько времени пользователь пишет код сам, а сколько — с помощью AI (ChatGPT, Claude, Copilot, Cursor, CLI-агенты), а также общую активность за компьютером.

Работает на **macOS** и **Windows** (с Linux в режиме installer-only без keystroke parity пока), покрывает **desktop apps**, **CLI** и **браузер**, отдаёт аналитику в дашборд.

## Quick start

- **Try as user:** download installer from [latest release](https://github.com/luckyrogue/eye-of-providence/releases/latest), register at `https://eop.rysdavletov.org`, pair device. See [`docs/alpha-install.md`](docs/alpha-install.md).
- **Self-host:** [`docs/self-hosting.md`](docs/self-hosting.md) — Docker Compose stack, single command.
- **Contribute:** [`CONTRIBUTING.md`](CONTRIBUTING.md) — repo layout, conventions, release flow.
- **Security:** [`.github/SECURITY.md`](.github/SECURITY.md) — disclosure policy + image-verification recipes.
- **Privacy:** [`docs/privacy.md`](docs/privacy.md) — what we collect, GDPR posture, data flow.
- **Operational risk register:** [`docs/tech-debt.md`](docs/tech-debt.md).

---

## 1. Что измеряем

### 1.1 Активность
- **Active time** — окно в фокусе + ввод за последние N сек (idle threshold 60–120с).
- **Idle / AFK** — нет input events, экран locked, screensaver.
- **Focus session** — непрерывный отрезок работы в одном приложении/проекте.

### 1.2 Coding time
- **Manual coding** — фокус в IDE/редакторе + активный keystroke stream + нет недавних AI-событий.
- **AI-assisted coding** — фокус × источник ввода (paste из AI-чата, Copilot inline accept, CLI tool call вернул код).
- **Reading / review** — фокус в редакторе без существенного ввода.

### 1.3 AI usage breakdown
- По провайдеру: OpenAI / Anthropic / Google / GitHub Copilot / Cursor / локальные модели.
- По каналу: browser chat, IDE inline, IDE chat panel, CLI agent, desktop app.
- По типу взаимодействия: chat, autocomplete, agent (multi-step), edit.

### 1.4 Code provenance (attribution)
Для каждого save/commit:
- **typed** — появилось через keystroke stream.
- **pasted-AI** — вставлено, clipboard заполнен после AI-источника.
- **pasted-other** — вставка без AI-источника.
- **AI-inline** — accept от Copilot/Cursor API.
- **AI-agent** — изменение от Claude Code / Cursor agent / Aider.

---

## 2. Архитектура

Унифицированный стек: **React + TypeScript** для всех UI, **Rust** для core агента (одна кодбаза, тонкие platform-modules), **Go** для бэкенда. 3 языка вместо 5.

```
┌────────────────────────────────────────────────────────────────┐
│                         CLIENT SIDE                            │
│                                                                │
│  ┌──────────────────────────────┐  ┌───────────────────────┐   │
│  │  Desktop App (Tauri 2)       │  │  Browser Extension    │   │
│  │  ┌─────────────────────────┐ │  │  (MV3, TS + React)    │   │
│  │  │ React UI                │ │  │                       │   │
│  │  │ (tray, settings, local  │ │  │  shared types & UI    │   │
│  │  │  dashboard view)        │ │  │  с desktop/web        │   │
│  │  └───────────┬─────────────┘ │  └──────────┬────────────┘   │
│  │              │ Tauri IPC     │             │                │
│  │  ┌───────────▼─────────────┐ │  ┌──────────▼────────────┐   │
│  │  │ Rust core               │ │  │  IDE / CLI plugins    │   │
│  │  │ • macOS: NSWorkspace,   │ │  │  • VS Code (TS)       │   │
│  │  │   CGEventTap            │ │  │  • JetBrains (V1)     │   │
│  │  │ • Windows: WinEventHook,│ │  │  • eop-hook (CLI)     │   │
│  │  │   GetLastInputInfo      │ │  └──────────┬────────────┘   │
│  │  │ • SQLite буфер, AES-GCM │ │             │                │
│  │  │   batch → ingest        │ │             │                │
│  │  └───────────┬─────────────┘ │             │                │
│  └──────────────┼───────────────┘             │                │
│                 └───────────────┬─────────────┘                │
└─────────────────────────────────┼──────────────────────────────┘
                                  │ HTTPS (JSON REST)
                                  ▼
┌────────────────────────────────────────────────────────────────┐
│                  BACKEND (Go monorepo)                         │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  cmd/api — unified HTTP API (auth, ingest, analytics, …)   │  │
│  └──────────────────────────┬───────────────────────────────┘  │
│  ┌──────────────────────────▼───────────────────────────────┐  │
│  │  cmd/worker — attribution post-processing (Phase A)      │  │
│  └──────────────────────────┬───────────────────────────────┘  │
│                             │                                  │
│  ┌──────────────────────────▼──────────────────────────────┐   │
│  │   Postgres (users/teams/devices)                        │   │
│  │   ClickHouse (events, sessions, attribution)            │   │
│  │   Redis (cache, WebAuthn challenges, rate limits)       │   │
│  └──────────────────────────┬──────────────────────────────┘   │
│                             │                                  │
│                  ┌──────────▼─────────┐                        │
│                  │  Web Dashboard      │  React SPA            │
│                  │  (тот же ui/ пакет, │  (shared с Tauri UI)  │
│                  │   что в Tauri)      │                       │
│                  └─────────────────────┘                       │
└────────────────────────────────────────────────────────────────┘
```

---

## 3. Компоненты — детали

### 3.1 Desktop App (Tauri 2, кросс-платформа)
Единое приложение для macOS и Windows. UI на React, native-доступ через Rust core.

**UI слой (React + TypeScript):**
- Tray icon + popup, окно настроек, локальный дашборд (read-only превью данных перед отправкой).
- Onboarding wizard с permissions-гайдом (скриншоты для Accessibility / Screen Recording).
- Pause/resume, blacklist apps & domains, экспорт/удаление данных.
- **Shared `ui/` пакет** — те же компоненты, что в web-дашборде и popup browser extension.

**Rust core (~1500 строк, кросс-платформенно с тонкими platform-modules):**
- `core/` — event schema (JSON, см. [`docs/api/openapi.yaml`](docs/api/openapi.yaml)), SQLite буфер (`rusqlite`), батчинг + retry + offline queue, AES-256-GCM (Keychain/DPAPI), redaction engine.
- `platform/macos/` — обёртки над `NSWorkspace.frontmostApplication`, `CGEventTap` (keystroke counts), `CGEventSourceSecondsSinceLastEventType` (idle), `NSPasteboard` (change count + хеши). Crates: `objc2`, `core-foundation`, `cocoa`.
- `platform/windows/` — обёртки над `SetWinEventHook(EVENT_SYSTEM_FOREGROUND)`, `GetForegroundWindow` + `QueryFullProcessImageName`, `GetLastInputInfo`, `SetWindowsHookEx(WH_KEYBOARD_LL)`, `GetClipboardSequenceNumber`. Crate: `windows-rs`.
- IPC: `tauri::command` для UI ↔ core, локальный HTTP `127.0.0.1:PORT` с токеном для browser extension и IDE plugins.

**Распространение:**
- Один кодбаз, один CI pipeline. Tauri собирает `.app` (signed + notarized) и `.msi` / `.exe` (signed) из одного `cargo tauri build`.
- Размер бинарника: 5–15 MB (vs 100+ MB у Electron).
- macOS: `launchd` agent (`~/Library/LaunchAgents/`). Windows: registry `Run` или Scheduled Task.

### 3.2 Browser Extension
- **Тип:** Manifest V3, Chrome / Edge / Firefox / Safari. **TypeScript + React** (popup использует тот же `ui/` пакет, что desktop и web).
- **Что детектит:**
  - AI-сайты: `chat.openai.com`, `chatgpt.com`, `claude.ai`, `gemini.google.com`, `perplexity.ai`, `copilot.microsoft.com`, `poe.com`, `you.com`, `cursor.sh`, `phind.com`, `bolt.new`, `v0.dev`, `lovable.dev`.
  - Время на вкладке в фокусе (`chrome.tabs.onActivated`, `chrome.windows.onFocusChanged`, visibility API).
  - Copy events на сообщениях ассистента (content script слушает `copy` на DOM узлах с message-ролью).
  - Длина скопированного, хеш — **не содержимое**.
- **Архитектура:** service worker (event hub) + content scripts per-domain (селекторы для каждого AI-сайта).
- **Связь с desktop agent:** локальный HTTP `127.0.0.1:PORT` с токеном, либо native messaging host.

### 3.3 IDE / CLI plugins
- **VS Code extension** (TypeScript):
  - События `vscode.workspace.onDidChangeTextDocument` — diff per save.
  - Copilot inline accept: hook на extension API (если доступен) или snapshot-сравнение.
  - Cursor: hook на их extension events.
  - Chat panel usage tracking.
- **JetBrains plugin** (Kotlin) — отложено до V1, через Platform SDK.
- **Claude Code:** `eop-hook` (`backend/cmd/eop-hook`) в `~/.claude/settings.json` на `Edit|Write|MultiEdit` → `POST /v1/ingest` (counts only). См. [`docs/attribution.md`](docs/attribution.md).
- **Commits:** `POST /v1/commits` (CI/script). Git `post-commit` hook — roadmap.
- **Generic CLI wrapper:** опц. shell-обёртка для `aider`, `gh copilot` — roadmap.

### 3.4 Web Dashboard
- **React SPA** на Vite, тот же `ui/` пакет, что в Tauri и browser extension.
- Чарты: Recharts / Visx. Стили: Tailwind.
- Раздаётся как статика с CDN; backend на Go отдаёт только API.

---

## 4. Attribution алгоритм

Цель: каждой строке/чанку кода в save/commit приписать категорию `typed | pasted-AI | pasted-other | AI-inline | AI-agent`.

### 4.1 Сигналы
- **Keystroke stream** с timestamp → ranges, набранные руками.
- **Paste events** + clipboard hash + clipboard source (последний AI-сайт/CLI, скопировавший этот hash в окне T мс назад).
- **IDE inline accept events** с координатами вставки.
- **Agent edit events** (Claude Code/Cursor пишет файл напрямую).

### 4.2 Pipeline
1. **Snapshot at open / save**: храним `(file, content_hash, ts)`.
2. **Diff на save**: hunks между предыдущим снимком и текущим.
3. **Match hunk → signal**:
   - Если hunk полностью покрыт интервалами keystrokes в этом файле → `typed`.
   - Если hunk совпадает с paste event и clipboard.source.is_ai → `pasted-AI` + provider.
   - Если hunk пришёл из IDE accept event → `AI-inline`.
   - Если hunk появился в момент, когда фокус был не в IDE (агент писал) → `AI-agent`.
   - Иначе → `unknown` (консервативно, лучше чем неверно).
4. **Time apportionment**: focus time на файле распределяется пропорционально размеру hunks × категории.
5. **Persist**: `attribution_event(user, file, project, lang, category, lines, chars, focus_ms, ts)`.

### 4.3 Edge cases
- Refactoring tools (rename, extract) → отдельная категория `refactor`, не путать с typed.
- Auto-format on save → исключаем из всех категорий.
- Generated files (`*.lock`, `dist/`) → blacklist.
- Bulk paste из документации (Stack Overflow) → `pasted-other`, не AI.
- Conflict resolution (merge) → отдельная категория.

---

## 5. Data model

### 5.1 Postgres (метаданные)
```sql
users(id, email, name, github_login, created_at, role, team_id, ...)
teams(id, name, plan, created_at, settings_json)
devices(id, user_id, os, hostname, fingerprint, last_seen)
projects(id, user_id, repo_url, root_path_hash, lang_primary)
api_tokens(id, user_id, scope, hashed_token, expires_at)
consent(user_id, scope, granted_at, revoked_at)
```

### 5.2 ClickHouse (события)
```sql
events (
  ts            DateTime64(3),
  user_id       UUID,
  device_id     UUID,
  session_id    UUID,
  app_bundle    LowCardinality(String),  -- com.microsoft.VSCode
  category      Enum8('idle','manual','ai','reading','other','refactor'),
  source        Enum8('os','browser','ide','cli'),
  ai_provider   LowCardinality(String),  -- 'openai','anthropic',...
  ai_channel    LowCardinality(String),  -- 'chat','inline','agent','cli'
  project_id    UUID,
  file_lang     LowCardinality(String),
  duration_ms   UInt32,
  chars_in      UInt32,                  -- набрано / вставлено
  lines_added   UInt32,
  lines_removed UInt32,
  meta          String                   -- JSON для гибкости
) ENGINE = MergeTree
ORDER BY (user_id, ts)
PARTITION BY toYYYYMM(ts)
TTL ts + INTERVAL 18 MONTH;
```

Materialized views для дешёвых агрегатов: daily/weekly per user × category × provider.

---

## 6. Auth & регистрация

### 6.1 OAuth провайдеры
- **GitHub** (приоритет — даёт язык/стек/репозитории сразу).
- **Google** (массовая аудитория).
- **Apple** (требуется для App Store если будем там).
- **Microsoft** (для корпоратов).

### 6.2 Onboarding flow
1. Sign up через OAuth → email, name, avatar.
2. Опросник (опц.): role (dev/PM/student/...), seniority, primary language, team size.
3. Скачивание агента + установка → пара "device ↔ user" через one-time code.
4. Permissions wizard (Accessibility, Screen Recording, browser extension, IDE plugin).
5. Первый дашборд показывается через 24ч активности (иначе пусто и фрустрирует).

### 6.3 "Кто это" — профилирование
- Авто-извлечение из GitHub: топ-языки, контрибьюты, org membership.
- Авто-определение стека по запущенным процессам (`node`, `python`, `rustc`, `go`, `cargo`, `dotnet`).
- Email domain → инференс компании / "personal" (gmail/yandex/proton).
- Cohorts: "новичок ≤2 года exp" / "senior" / "lead", "frontend / backend / ML / devops", "solo / team / enterprise".

---

## 7. Dashboards & analytics

### 7.1 Personal
- Heatmap активности (по часам/дням).
- AI ratio: % времени с AI vs manual, тренд за 7/30/90 дней.
- Top AI providers / channels.
- Per-project breakdown.
- Per-language breakdown ("на Rust ты пишешь сам, на TS — 70% AI").
- "Dependency score" — насколько растёт зависимость от AI.
- Streaks, weekly recap email.

### 7.2 Team / admin
- Aggregated only (нельзя смотреть чужие сессии без opt-in).
- Adoption rate: сколько % команды реально используют AI.
- ROI proxy: AI ratio × shipped commits.
- Onboarding indicator: новые члены команды и кривая их AI-усыновления.
- Tool consolidation insights: "5 человек платят за Cursor, 3 за Copilot — стоит ли унифицировать".

### 7.3 Product analytics (для нас)
- Acquisition: source → activated → retained (D1/D7/D30).
- Funnel: signup → install agent → first event → 7d active.
- Cohort retention by acquisition channel.
- Feature usage: какие интеграции реально включены.

### 7.4 AI-generated reports (Gemini)
В конце недели / месяца / спринта пользователь может сгенерировать **narrative-отчёт** на основе своих метрик. Цифры превращаются в читаемое summary с инсайтами, паттернами и рекомендациями.

**Провайдер:** Google **Gemini** (`gemini-2.5-flash` для regular reports, `gemini-2.5-pro` для глубоких квартальных).

**Что попадает в отчёт:**
- Топ-инсайт ("на этой неделе ты на 40% больше использовал AI в TypeScript-проектах vs прошлой").
- Изменения в паттернах (рост/падение AI-ratio, новые провайдеры, всплески idle).
- Сравнение с собственными baseline и (опц.) с anonymized cohort ("разработчики с твоим стажем в TS используют AI в среднем 55%, ты — 70%").
- Конкретные рекомендации ("ты тратишь 2ч/день на ChatGPT в браузере — рассмотри Cursor/Claude Code, который интегрирован в IDE").
- Highlights / lowlights недели.

**Архитектура:**
- Триггер: cron (weekly/monthly) или ручная кнопка "Generate report" в дашборде.
- Go worker берёт агрегаты из ClickHouse → формирует **только числовой контекст** (никакого кода, промптов, имён файлов в отчёт) → шлёт в Gemini API.
- Промпт-шаблоны хранятся в `backend/internal/reports/prompts/` (system prompt + few-shot examples под каждый тип отчёта).
- Ответ сохраняется в Postgres как `reports(id, user_id, period, format, body_md, generated_at, model)`.
- Дашборд рендерит markdown + инлайн чарты из тех же агрегатов.

**Privacy при отправке в Gemini:**
- Шлём **только агрегированные числа** и метки (язык, провайдер, категория) — никогда содержимое файлов / промптов / clipboard / window titles.
- User ID не пересылается, только локальный hash отчёта.
- Opt-in: AI-reports отключены по умолчанию, включаются явно в settings (с экраном "что именно отправляется в Google").
- Альтернатива для приватных пользователей: локальный режим через Ollama (`gemma`/`llama3`) — тот же промпт, без отправки наружу. План на V1.

**Форматы:**
- Personal weekly digest (markdown в дашборде + email).
- Monthly retrospective.
- Team report (только из агрегатов команды, доступен admin).
- Quarterly career-style report ("за квартал ты освоил X, AI-зависимость в Y снизилась").

**Cost-control:**
- Gemini Flash дёшев (~$0.075/1M input tokens), агрегаты компактные → один отчёт ~5–15K input tokens.
- Кэширование промптовой части (system + few-shot) через Gemini context caching.
- Rate limit: 1 weekly report / user / неделя без оплаты, далее — paid plan.

---

## 8. Privacy & trust

**Принципы (некомпромиссные):**
1. **Local-first**: вся обработка на клиенте, на сервер уходят только агрегаты + метаданные событий.
2. **Никогда не отправляем:**
   - Содержимое файлов / промптов / ответов AI.
   - Сами keystrokes (только counts).
   - Заголовки окон в private/incognito или из blacklist (банкинг, мессенджеры).
   - Содержимое clipboard (только хеш + размер).
3. **Open-source клиент** — код агента и расширения публичны, можно self-host backend.
4. **Local-only mode** — solo-пользователь может вообще не подключать backend.
5. **Team mode = explicit opt-in**, admin видит только агрегаты по команде, не индивидуальные сессии. Это прописывается на уровне consent в onboarding и enforced на бэкенде.
6. **Данные deletable**: один клик "delete all my data" → реально стирает из CH/PG.
7. **Aggregation thresholds**: метрики команды показываются только если ≥5 активных людей (k-anonymity).
8. **Audit log** — кто из админов что смотрел.

**Compliance:** GDPR (DPA, право на забвение, экспорт), SOC2 — план на post-MVP.

---

## 9. Roadmap

### MVP (6 недель)
- **W1–2:** macOS native agent + локальный SQLite + базовая категоризация app→category. Onboarding-скрипт с permissions.
- **W2–3:** Chrome extension для AI-сайтов. Связка с агентом через native messaging.
- **W3–4:** Backend: auth (GitHub OAuth), ingest API, минимальный CH-схема, Vite + React дашборд с personal-метриками.
- **W4–5:** Windows agent (паритет с macOS на window/idle/keystroke counts).
- **W5–6:** Attribution v1 (clipboard-based) + Claude Code hooks + VS Code extension для accept-events.

### V1 (3 месяца)
- JetBrains plugin.
- Cursor/Copilot/Aider deeper integration.
- Team mode + admin dashboard.
- Weekly recap email.
- **AI-generated reports через Gemini** (weekly/monthly digest с инсайтами).
- Self-host docker-compose.

### V2 (6 месяцев)
- ML-attribution (улучшить unknown bucket через классификатор по шаблонам кода).
- Mobile app для просмотра дашборда.
- Slack/Discord integration (weekly digest в канал).
- Локальный AI-reports через Ollama (для privacy-purists).
- Public API + webhooks.
- SOC2 type 1.

---

## 10. Risks & mitigations

| Риск | Митигейшн |
|------|-----------|
| macOS Accessibility permission — высокий churn на onboarding | Подробный wizard с скриншотами, fallback-режим (window-level only, без keystrokes) |
| False positives в attribution | Консервативный классификатор, видимая категория `unknown`, возможность пользователю переразметить |
| Восприятие "слежки" → отказ ставить | OSS клиент, local-only mode, явные privacy guarantees, нет content collection |
| Разная точность на двух OS | Shared core в Rust, контракт-тесты на event schema |
| Team-режим как инструмент микроменеджмента | Hard-cap на гранулярность (агрегаты, k-anonymity, нет per-session view для админа) |
| Anti-cheat в team-режиме | Clock-skew detection, hash-chain событий, device attestation |
| Cross-device dedup (лаптоп + десктоп = 1 человек) | Device fingerprint + user_id linking, conflict resolution в ingest |
| Скейл ClickHouse при росте | Партиционирование по месяцу, TTL 18 мес, materialized views для горячих запросов |
| AI-reports утекают данные в Gemini | Шлём только агрегаты-числа, opt-in, явный экран "что отправляется", local fallback (Ollama) для приватных |
| Gemini-косты при росте | Flash-модель, context caching, rate limit на free tier, бэкграунд cron вместо real-time |

---

## 11. Tech stack (итог)

3 основных языка: **TypeScript / React** для всех UI, **Rust** для core агента, **Go** для бэкенда. Kotlin — опц. для JetBrains plugin (V1).

- **Desktop agent:** Tauri 2 + Rust core (`tokio`, `rusqlite`) с тонкими platform-modules (`objc2`/`core-foundation` для macOS, `windows-rs` для Windows). UI на React + TypeScript + **shadcn/ui**.
- **Browser extension:** TypeScript + React + **shadcn/ui**, Manifest V3.
- **IDE plugins:** TypeScript (VS Code). Kotlin для JetBrains — отложено.
- **CLI:** `eop-hook` binary для Claude Code (`backend/cmd/eop-hook`).
- **Backend:** Go + **Fiber** (JSON REST), **zap**, `pgx`, ClickHouse Go client. Postgres, ClickHouse, Redis.
- **AI reports:** **Google Gemini API** (`gemini-2.5-flash` / `pro`) через официальный Go SDK (`google.golang.org/genai`), с context caching. Опц. локальный fallback через Ollama.
- **Web dashboard:** React SPA на Vite, **shadcn/ui** (Radix UI + Tailwind), Recharts/Visx. Тот же `ui/` пакет, что и в Tauri / browser extension.
- **Shared event contract:** JSON schema в [`docs/api/openapi.yaml`](docs/api/openapi.yaml) (`Event`); ingest `POST /v1/ingest`.
- **Infra:** Docker, Kubernetes (или Fly.io на старте), Terraform, GitHub Actions.
- **Observability:** OpenTelemetry, Grafana, Sentry.

---

## 12. Repo layout (план)

```
eye-of-providence/
├── agent/                  # Tauri 2 desktop app (macOS + Windows из одной кодбазы)
│   ├── src/                # React + TS UI (tray, settings, local dashboard)
│   └── src-tauri/          # Rust core
│       ├── core/           # event schema, SQLite, batching, encryption
│       ├── platform/macos/ # NSWorkspace, CGEventTap, NSPasteboard
│       └── platform/windows/ # WinEventHook, GetLastInputInfo, clipboard
├── browser-extension/      # MV3, TS + React (использует ui/)
├── ide-vscode/             # TS extension
├── backend/                # Go monorepo (Fiber + zap)
│   ├── cmd/
│   │   ├── api/            # production HTTP API
│   │   ├── worker/         # attribution post-processing
│   │   ├── eop-hook/       # Claude Code ingest hook
│   │   ├── migrate/        # standalone migrations CLI
│   │   └── reports/        # AI reports cron (dev/ops)
│   ├── internal/
│   │   └── migrate/sql/    # Postgres + ClickHouse migrations
│   └── …
├── dashboard/              # React SPA (Vite), переиспользует ui/
├── ui/                     # shared React components на shadcn/ui
├── infra/                  # docker-compose для self-host
├── docs/                   # см. docs/README.md
└── README.md
```

---

## 13. Open questions

- Free tier vs paid? (Гипотеза: free для solo, paid для team starting from $X/seat.)
- Self-host edition — open-core или fully open?
- Стоит ли с самого начала собирать prompt embeddings для AI-quality метрик, или это сразу убьёт privacy-позиционирование?
- Mobile companion app — нужен или distraction?
