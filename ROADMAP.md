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
- ✅ Агенты — code-side hardened (MV3 race fix, Tauri pause+GC, VSCode multi-window dedup, retry queues). Distribution в marketplace остаётся внешним блокером.
- ⚠️ Attribution алгоритм — primitive (paste size + AI domains)
- ✅ Тесты (backend 42.5%), observability (logs+histograms), transactional emails (Resend)
- ❌ Billing — отложен до выхода из беты

**Главные риски на год:**
1. Агенты не работают надёжно → дашборд пустой → нет ценности → нет апгрейда
2. Не нашли ICP — eng managers vs developers? Решается на 3 founding companies в Q1
3. AI-attribution-тема может устареть, если все юзают AI. Pivot в "code quality + AI accountability"
4. Cursor/Copilot выпустят свою аналитику — наш angle: independent / multi-vendor (видим всё, они только своё)

---

## Q1 — Foundation (июнь-август 2026)

> **Тема: продукт перестаёт ломаться без оператора.**
>
> **Статус (2026-05-10): закрыт раньше срока, осталось 2 пункта. Можно переходить к Q2.**

### Фичи (пользовательский путь)

- ✅ **Агенты до production-quality** — code-side готов:
  - ✅ Tauri: pause flag (UI+tray), SQLite GC (>7 дней), idle-detection, retry queue
  - ✅ VS Code: multi-window dedup (focused-window-only)
  - ✅ Browser extension: MV3 service-worker state в chrome.storage.session, retry queue с exponential backoff
  - ⚠️ **Distribution в marketplace** — внешний блокер. См. `docs/agents-publishing.md`. Требует Apple Dev $99/год, Chrome Web Store $5, MS Partner / VS Code Azure DevOps publisher.
- ✅ **Onboarding wizard** после регистрации (4 шага: company → install → invite → event)
- ✅ **Transactional emails** (Resend): invite, password reset
- ✅ **Constraint: 1 owner = 1 company** через `pg_advisory_xact_lock` + 1-owner-per-user invariant
- ~~Billing scaffolding через Stripe~~ — **отложено до выхода из беты**. Пока остаёмся полностью free для founding companies.

### Tech debt

- ✅ **Тесты**: backend coverage **42.5%** (цель была 40%). 30+ integration-тестов на teams/auth/admin/password-reset/members/projects/commits/invites.
- ❌ **E2E Playwright** на dashboard (login → dashboard → invite → second user joins) — единственный незакрытый Q1-пункт.
- ✅ **Observability**: zap structured logs в stdout, /metrics с histograms (request + CH read/write), Uptime Kuma docs. Sentry отложен.
- ✅ **DB миграции с down-сценариями**: переход на golang-migrate с pgx/v5 driver, все миграции имеют up+down.

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

- ✅ **Refactor backend в clean packages** (сделано раньше срока): `teams/handler.go` 1397 LoC → 8 доменных файлов (auth/password_reset/teams/members/invites/projects/commits/admin).
- ✅ **Frontend: react-query + react-router** (сделано раньше срока): @tanstack/react-query 5.x, react-router-dom 6.30, FSD-структура (app/pages/widgets/entities/shared).
- ❌ **OpenAPI spec + типизированный клиент** для frontend (`orval` / `openapi-typescript`). Сейчас руками поддерживается `entities/*/api/*.ts` — будет источник багов.
- ❌ **ClickHouse query benchmarks + индексы.** На 10M событий в день текущие запросы сложатся. Профилировать сейчас, не когда сложится. Materialized views для daily/weekly aggregates.

### Метрика квартала

Activation rate > 70%, week-2 retention > 40%, ≥1 paying customer.

---

## Q3 — Market (декабрь-февраль 2026/27)

> **Тема: продукт обрастает аудиторией. Pricing — отдельный решительный момент, но не дефолт.**

### Фичи

- ~~Pricing live~~ — **остаёмся free, beta-лимит 3 пока сохраняется.** Решение о монетизации принимаем по сигналам:
  - Если 3 founding companies активно пользуются + есть waitlist > 30 — снимаем лимит, делаем self-serve free signup, отдельно начинаем разговор о Pro-tier.
  - Если активность слабая — продолжаем итерации с founding companies, **не отвлекаемся на billing**.
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

WAU > 30 (3 компании × ~10 активных), week-2 retention > 50%, активный pipeline у каждой founding company.

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
| **Privacy/Security** | ✅ secrets audit, rate limiting, CodeQL/trivy/gitleaks SAST, 4 алерта закрыты | Penetration test (3rd party), Sentry / similar | SOC2 prep + Vanta | SOC2 Type 1 audit |
| **Документация** | ✅ docs/agents-publishing.md, ✅ Uptime Kuma setup. README + setup guides — частично | API docs (OpenAPI auto-gen) | User docs site (mintlify) | Enterprise / SOC2 docs |
| **i18n** | ✅ RU/EN/KK/ES (433 keys × 4 локали, 100% parity) | — | — | — |
| **Команда** | соло | +1 frontend (contractor or full-time) | +1 backend | +1 DevRel/marketing |
| **Метрики** | DAU, signups, activation | MRR (≥1), retention W2/W4 | MRR ≥ $1k, conversion ≥ 5% | MRR ≥ $5k, NRR ≥ 110% |

---

## Tech debt long list (горизонт 12+ мес)

В порядке убывания приоритета:

1. ✅ **Test coverage**: 0% → **42.5%** (Q1, цель 40% выполнена) → 60% (Q2) → 70% (Q3+)
2. **Observability**: ✅ stdout+histograms+Uptime Kuma (Q1) → Sentry + Grafana (Q2) → SLO dashboards (Q3)
3. **CI/CD**: ✅ reusable workflows + path-filters + SAST (Q1) → staging + preview deploys (Q3)
4. ✅ **DB migrations**: only up → up+down (Q1, golang-migrate с pgx/v5) → schema versioning (Q4)
5. **Secrets**: `.env` → Doppler/Vault (Q3)
6. **Error handling**: ad-hoc → typed errors + structured response (Q2)
7. ✅ **Frontend**: vanilla useState → react-query + react-router-dom + react-hook-form (Q1, FSD structure)
8. **No event sourcing/audit trail** для team actions (Q4)
9. **ClickHouse perf**: not benchmarked → materialized views + partitioning (Q2-Q4)
10. **No mobile** (Q3)
11. **Single region** (Q4)
12. **Manual deploy of agents**: Tauri autoupdate + signed builds — ⚠️ release pipeline есть, signing требует Apple/MS аккаунтов (Q1)

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
- ❌ Mobile native attribution (iOS / Android coding-tracking — слишком nichе)
- ❌ Code review automation (это competitive с Greptile/Coderabbit, не наша битва)
- ❌ Свой billing engine (Stripe всегда)
- ❌ Кастомный dashboarding для customers (BI-export через API)

> ~~Локализация (только RU + EN)~~ — изменено: i18n охватывает RU/EN/KK/ES (4 локали × 433 ключа, 100% parity).

---

_Последнее обновление: 2026-05-10. Ревью раз в квартал._

---

## Q1 retro (на 2026-05-10, до формального старта Q1)

Quarter ahead-of-schedule. Из 11 крупных пунктов Q1 закрыто 9, два (E2E Playwright, marketplace distribution) висят. Дополнительно сделано вне плана:

- i18n × 4 локали (RU/EN/KK/ES) — изначально планировался RU+EN
- Backend refactor (был Q2)
- Frontend на react-query + react-router (был Q3)
- Tech-lead CI pipeline (reusable workflows + path-filters + SAST/secret/CVE сканеры)

**Backlog Q2 (по приоритету):**
1. Attribution v2 — Copilot Accept event, Cursor Apply, Claude Code hook (real moat)
2. Insights, не данные — narrative-вместо-графиков
3. Public API + webhooks — снимает страх vendor lock-in
4. OpenAPI spec + типизированный клиент — последний source-of-bugs в frontend
5. ClickHouse benchmarks + materialized views — pre-emptive перед 10M событий/день
6. Slack integration — самый дешёвый retention loop
7. Sentry-or-similar — first paying customer trigger
8. Penetration test (3rd party) — pre-Q3 SOC2 prep
