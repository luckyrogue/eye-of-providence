# Roadmap — Eye of Providence

Горизонт: 12 месяцев с **июня 2026 по май 2027**. Команда: 1-3 человека.

Каждый квартал — фичи + tech debt параллельно (~60/40), плюс одна сквозная тема. Цель — из beta-продукта с 3 founding companies дойти до **enterprise-ready** SaaS с MRR.

---

## Состояние на старт (май 2026)

- ✅ Backend (Go), Postgres + ClickHouse + Redis, Dokploy на Hetzner-VM
- ✅ Dashboard (React/Vite), CI с GHCR push
- ✅ Multi-tenant team management, invite-flow, RBAC
- ✅ Beta-лимит 3 компании
- ✅ Landing page
- ⚠️ Агенты (Tauri / VS Code / browser ext) — на 30-60% готовности
- ⚠️ Attribution алгоритм — primitive (paste size + AI domains)
- ❌ Тесты, observability, billing, transactional emails

**Главные риски на год:**
1. Агенты не работают надёжно → дашборд пустой → нет ценности → нет апгрейда
2. Не нашли ICP — eng managers vs developers? Решается на 3 founding companies в Q1
3. AI-attribution-тема может устареть, если все юзают AI. Pivot в "code quality + AI accountability"
4. Cursor/Copilot выпустят свою аналитику — наш angle: independent / multi-vendor (видим всё, они только своё)

---

## Q1 — Foundation (июнь-август 2026)

> **Тема: продукт перестаёт ломаться без оператора.**

### Фичи (пользовательский путь)

- **Агенты до production-quality** — самый важный пункт квартала.
  - Tauri desktop: macOS code-signing + notarization, Windows MSI, autoupdate, надёжный idle-detection, retry queue для оффлайна.
  - VS Code extension: публикация в marketplace, накапливание Copilot Accept events, корректная обработка multi-window.
  - Browser extension: Chrome Web Store + Firefox AMO, fix MV3 background-worker гонок.
  - 2-3 недели на каждый. **Без этого дашборд пустой.**
- **Onboarding wizard** после регистрации: 4 шага «Создай команду → установи агент → пригласи людей → отправь первое событие». Сейчас новый owner видит белый экран.
- **Transactional emails** (Resend / Postmark): invite, password reset, weekly report digest, alert на инвайт-лимит. Must-have для team-продукта.
- **Billing scaffolding** через Stripe: даже на free — Customer / Subscription объекты в БД, чтобы переход на paid не требовал миграции данных.

### Tech debt

- **Тесты**. Сейчас покрытие околонулевое. Цель Q1: **40% line coverage в `backend/`**.
  - Integration-тесты на auth flow, ingest pipeline, team RBAC, beta-limit.
  - E2E на dashboard через Playwright (login → dashboard → invite → second user joins).
- **Observability**. Sentry frontend + backend, structured logs в Loki/Grafana cloud free tier, Uptime Kuma для public status. Без этого первый incident у paying customer поломает доверие.
- **DB миграции с down-сценариями**. Сейчас только up. Для production rollback — блокер.

### Метрика квартала

DAU > 10, activation rate (signup → первый event в течение 24h) > 50%.

---

## Q2 — Value depth (сентябрь-ноябрь 2026)

> **Тема: продукт даёт инсайты, за которые платят.**

### Фичи

- **Attribution v2.** Текущая логика (paste size + AI-домены) даёт ~70% точности. Добавить:
  - Copilot Accept event (есть API в VS Code)
  - Cursor Apply event
  - Claude Code hook → реальный multi-step agent attribution
  - Цель: **точность 90%+**. Это и есть наш реальный moat.
- **Insights, не данные**: «Ты ускорился на 23% в Python за месяц», «Команда тратит 40% времени на review AI-кода», «Этот PR на 80% от AI». Сейчас графики сырые. Гемини-отчёты раз в неделю — мало.
- **Public API + webhooks.** «Ваши данные ваши» → экспорт, BI-tools, кастомные дашборды. Снимает страх vendor lock-in.
- **Slack integration**: weekly digest в канал команды. Самый дешёвый retention-loop.

### Tech debt

- **Refactor backend в clean packages.** `teams/handler.go` сейчас 800 строк, всё в одном файле. Разбить на `Service` / `Repo` / `HTTPHandler`. Разлочит парную разработку.
- **OpenAPI spec + типизированный клиент** для frontend (`orval` / `openapi-typescript`). Сейчас руками поддерживается `api.ts` — будет источник багов.
- **ClickHouse query benchmarks + индексы.** На 10M событий в день текущие запросы сложатся. Профилировать сейчас, не когда сложится. Materialized views для daily/weekly aggregates.
- **Frontend: react-query + react-router.** Сейчас локальные `useState` + кастомный роутер по `pathname`. На масштабе придётся переписывать.

### Метрика квартала

Activation rate > 70%, week-2 retention > 40%, ≥1 paying customer.

---

## Q3 — Market (декабрь-февраль 2026/27)

> **Тема: продукт продаётся.**

### Фичи

- **Pricing live.** Free → $X/seat. Stripe billing, invoice generation, dunning emails. Снять beta-лимит 3 компании.
- **Marketing**:
  - Blog (1 пост в неделю): технические разборы attribution, истории команд, бенчмарки AI-coders
  - Changelog page (auto-generated из conventional commits)
  - Public roadmap (open vote на фичи, типа productlane / canny)
  - Case studies от первых 3 founding companies
- **Integrations**:
  - GitHub deeper: PR-комментарий «AI added 80% of this PR», hooks на merge для автоматического attribution
  - Linear / Jira: связка коммитов с тикетами
  - Slack alerts: «команда сегодня необычно много AI», «деплой за пятничный вечер»
- **Mobile companion (read-only)**: iOS / Android, дашборд + еженедельный push с отчётом. Retention, не acquisition.

### Tech debt

- **CI/CD maturity**:
  - Staging environment с feature-flags
  - Preview deployments per PR (Dokploy supports this)
  - Smoke tests после деплоя
  - Blue-green deploy для API, чтобы апдейт не ломал WebSocket-соединения от агентов
- **Secret management**: `.env` в Dokploy → Doppler / 1Password Connect / Vault. Когда команда вырастет — без этого не масштабироваться.
- **Frontend архитектура переезд** на нормальные библиотеки:
  - `@tanstack/react-router` или `wouter` (легковесный)
  - `@tanstack/react-query` для кэша
  - `sonner` для toast-системы
  - `zod` для валидации форм
  - Это разлочит скорость разработки на ×2.

### Метрика квартала

MRR > $1k, signup-to-paid conversion > 5%, churn < 5%/mo.

---

## Q4 — Scale (март-май 2027)

> **Тема: продукт держит enterprise customer.**

### Фичи

- **SSO (SAML / OIDC)**. Enterprise sales hard requirement. Без этого выше **$500/mo** не продать.
- **Audit log**: кто что сделал в команде. Compliance + trust.
- **SOC2 path**: Vanta / Drata, security policies, ежегодный pentest. Не сертификат, а **«в процессе»** — этого хватает на $5k MRR клиентов.
- **Self-hosted enterprise edition**: airgapped install, Helm chart, license keys, премиум-поддержка.

### Tech debt

- **Performance audit**:
  - ClickHouse partitioning по `customer_id` (или sharding если очень много)
  - Materialized views для дашбордов (раз в час пересчёт, дашборд читает агрегаты)
  - Redis кэш на hot queries (top languages, weekly summary)
- **Disaster recovery**:
  - Backups Postgres + ClickHouse в S3 (отдельный регион)
  - Runbook с RTO < 4h, RPO < 1h
  - **Ежемесячные restore drills** — не «настроили и забыли»
- **Multi-region readiness** (не deploy, но архитектурно готов): customer_id → region binding в схеме. Без этого первый европейский enterprise клиент с GDPR завернёт.

### Метрика квартала

MRR > $5k, NRR > 110%, ≥1 enterprise (>$500/mo) customer.

---

## Сквозные темы

| Тема | Q1 | Q2 | Q3 | Q4 |
|---|---|---|---|---|
| **Privacy/Security** | Sentry, secrets audit, rate limiting | Penetration test (3rd party) | SOC2 prep + Vanta | SOC2 Type 1 audit |
| **Документация** | README + setup + agent install guides | API docs (OpenAPI auto-gen) | User docs site (mintlify) | Enterprise / SOC2 docs |
| **Команда** | соло | +1 frontend (contractor or full-time) | +1 backend | +1 DevRel/marketing |
| **Метрики** | DAU, signups, activation | MRR (≥1), retention W2/W4 | MRR ≥ $1k, conversion ≥ 5% | MRR ≥ $5k, NRR ≥ 110% |

---

## Tech debt long list (горизонт 12+ мес)

В порядке убывания приоритета:

1. **Test coverage**: 0% → 40% (Q1) → 60% (Q2) → 70% (Q3+)
2. **Observability**: нет → Sentry + Grafana (Q1) → SLO dashboards (Q3)
3. **CI/CD**: одиночный pipeline → staging + preview deploys (Q3)
4. **DB migrations**: only up → up+down (Q1) → schema versioning (Q4)
5. **Secrets**: `.env` → Doppler/Vault (Q3)
6. **Error handling**: ad-hoc → typed errors + structured response (Q2)
7. **Frontend**: vanilla useState → react-query/router/zod (Q3)
8. **No event sourcing/audit trail** для team actions (Q4)
9. **ClickHouse perf**: not benchmarked → materialized views + partitioning (Q2-Q4)
10. **No mobile** (Q3)
11. **Single region** (Q4)
12. **Manual deploy of agents**: Tauri autoupdate + signed builds (Q1)

---

## Pivot triggers

Если что-то из списка ниже произойдёт — переоценка roadmap.

- **Cursor/Copilot выпускают свою аналитику** → форсировать «multi-vendor independent» позиционирование, отказаться от race на features
- **Не находим ICP за Q1** → pivot в индивидуальный продукт (B2C для разработчиков) или закрыть проект
- **AI-кодинг становится нормой** → расширить scope с «attribution» на «AI-augmented engineering productivity» (включая review, тесты, security)
- **Cloudflare/Tailscale покупает конкурента** → ускоренный exit или cooperation

---

## Не делаем в этом году

Чтобы не размывать фокус, **явно отказываемся** от:

- ❌ Свой LLM / generative features (Gemini хватает)
- ❌ On-prem deployment с поддержкой kubernetes-versions <1.28
- ❌ Локализация (только русский + английский)
- ❌ Mobile native attribution (iOS / Android coding-tracking — слишком nichе)
- ❌ Code review automation (это competitive с Greptile/Coderabbit, не наша битва)
- ❌ Свой billing engine (Stripe всегда)
- ❌ Кастомный dashboarding для customers (BI-export через API)

---

_Последнее обновление: 2026-05-08. Ревью раз в квартал._
