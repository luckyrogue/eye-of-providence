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
>
> **Статус (2026-05-10): закрыт раньше срока 6/8. Sentry+pentest отложены per design.**

### Фичи

- ✅ **Attribution v2** (commit `30fbedf`): Cursor detection через `vscode.env.appName`, Copilot inline burst-heuristic (несколько contentChanges за <100ms), Claude Code PostToolUse hook (`backend/cmd/eop-hook`). 7 unit-тестов. Цель 90%+ accuracy достигнута для Claude-Code path (zero ambiguity); Cursor/Copilot — best-effort burst.
- ✅ **Insights, не данные** (commit `8cebbd2`): 5 narrative-карточек (ai_trend up/down/flat/started, ai_ratio, top_lang, productive_day, total_activity) с Z-score noise floor. Backend: `internal/insights/`. Frontend: `widgets/insights`. 14 unit-тестов.
- ✅ **Public API + webhooks** (3 commits `027b79f`+`2f85324`+`e2e04fd`): API tokens (eop_*) с scopes (read/write:ingest/admin), HMAC-signed webhooks с retry + Slack format adapter (commit `264b5a0`), CRUD UI в Settings, OpenAPI 3.1 spec + TS-types generation.
- ✅ **Slack integration** (commit `264b5a0`): payload format adapter для Slack incoming-webhook URLs (Block Kit). Anomaly alerts daily cron (`internal/anomaly/`) с Z-score detection (commit `2f56c13`).

### Tech debt

- ✅ **Refactor backend в clean packages** (сделано раньше срока): `teams/handler.go` 1397 LoC → 8 доменных файлов (auth/password_reset/teams/members/invites/projects/commits/admin).
- ✅ **Frontend: react-query + react-router** (сделано раньше срока): @tanstack/react-query 5.x, react-router-dom 6.30, FSD-структура (app/pages/widgets/entities/shared).
- ✅ **OpenAPI spec + типизированный клиент** (commit `e2e04fd`): `docs/api/openapi.yaml` 3.1 + `openapi-typescript@7` → 905 LoC сгенерированных TS-types в `shared/api/openapi.d.ts`.
- ✅ **ClickHouse benchmarks + materialized views** (commit `576e01c`): `events_hourly_agg` SummingMergeTree + MV catches inserts. **3-7× speedup** на hot queries (Aggregate/DailyTrend/LanguageBreakdown/Heatmap), 35× storage reduction. Benchmark CLI `cmd/ch-bench`.

### Метрика квартала

Activation rate > 70%, week-2 retention > 40%, ≥1 paying customer.

---

## Q3 — Market (декабрь-февраль 2026/27)

> **Тема: продукт обрастает аудиторией. Pricing — отдельный решительный момент, но не дефолт.**
>
> **Статус (2026-05-10): закрыт раньше срока 8/11. Только marketing-content (Blog, Public roadmap, Case studies) + external infra setup (CI/CD staging, Secrets vault) висят.**

### Фичи

- ~~Pricing live~~ — **остаёмся free, beta-лимит 3 пока сохраняется.** Решение о монетизации принимаем по сигналам:
  - Если 3 founding companies активно пользуются + есть waitlist > 30 — снимаем лимит, делаем self-serve free signup, отдельно начинаем разговор о Pro-tier.
  - Если активность слабая — продолжаем итерации с founding companies, **не отвлекаемся на billing**.
- **Marketing**:
  - ❌ Blog (1 пост в неделю) — content-работа, не code
  - ✅ **Changelog page** (commit `53deef8`): public `/changelog` route, auto-generated из conventional commits через `cmd/changelog-gen` → `dashboard/public/changelog.json`. 12×4 i18n keys.
  - ❌ Public roadmap (open vote) — external service (productlane/canny) либо самопис
  - ❌ Case studies — content-работа, ждёт активность founding companies
- **Integrations**:
  - ✅ **GitHub PR-comment + GitLab MR** (commit `f1b7d13`): pull-based endpoint `/v1/integrations/pr-comment` aggregates AI% по commits, форвардит в GH/GL API через user'ский PAT. Markdown comment с progress-bar + breakdown. Privacy: provider_token не сохраняется. 15 unit + http-mock тестов.
  - ✅ **ClickUp linking** (commit `2f56c13`, переориентирован с Linear/Jira per user feedback): regex `CU-<id>` + `#<id>` в commit messages → clickable `app.clickup.com/t/{id}` links в commits-tab. 14 inline tests.
  - ✅ **Slack anomaly alerts** (commit `2f56c13`): daily Z-score cron, Block Kit rendering per kind (rocket/turtle/hammer/sparkles). Shared с PWA push delivery (commit `306652c`).
- ✅ **Mobile companion (PWA)** (commits `82b0dc0`+`306652c`+`671b6a9`): installable PWA shell (manifest, service worker, offline cache, install prompts iOS+Chromium), Web Push subscriptions с VAPID, native-app-style adaptive layout с bottom-nav, safe-area insets, 44×44 tap targets. Anomaly cron шлёт push параллельно со Slack.

### Tech debt

- ❌ **CI/CD maturity** (Dokploy access required):
  - Staging environment с feature-flags
  - Preview deployments per PR
  - Smoke tests после деплоя
  - Blue-green deploy
- ❌ **Secret management**: `.env` → Doppler / 1Password Connect / Vault — нужен external setup
- ✅ **Frontend архитектура переезд** (закрыт полностью):
  - ✅ react-query@5 для кэша (Q1)
  - ✅ react-router-dom@6.30 (Q1)
  - ✅ sonner для toast (Q1)
  - ✅ **zod** для валидации форм (commit `e4a0fd1`): centralized schemas в `shared/lib/schemas.ts`, react-hook-form через zodResolver, error messages — i18n keys
- ✅ **Внеплановая инфра-чистка** (вне Q3-плана):
  - **Caddy migration** (commit `8e91ca3`): nginx → caddy:2.11 в production image, env-vars нативно через `{$VAR:default}` без envsubst step
  - **Trivy CVE waivers** (commit `0b79011`): `.trivyignore` для 2 caddy CVE которые not-exploitable в нашей config (auto_https off + нет gRPC endpoints)
  - **shadcn/Radix Select** (commit `2b96d6d`): replace native `<select>` на portal-popover Select с `SimpleSelect` helper (data-driven options) — 6 widgets migrated
  - **Comprehensive mobile audit** (commits `8023722`+`7de8de4`): 15 файлов responsive pass (`px-6 → px-4 sm:px-6`, heading scale chains, hero CTAs flex-col stack, ProductPreview tile shrinks, card-headers stack vertically)

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
| **Privacy/Security** | ✅ secrets audit, rate limiting, CodeQL/trivy/gitleaks SAST, 4 алерта закрыты, .trivyignore с justification | Penetration test (3rd party) — отложен; Sentry — отложен per design (до paying) | ✅ HMAC webhooks, encodeURIComponent CSRF prevention, OpenAPI security schemas | SOC2 Type 1 audit |
| **Документация** | ✅ docs/agents-publishing.md, ✅ Uptime Kuma setup. README + setup guides — частично | ✅ OpenAPI 3.1 spec в docs/api/openapi.yaml + auto-generated TS types | ✅ docs/clickhouse-perf.md, docs/integrations-pr-comment.md, browser-extension/README.md | Enterprise / SOC2 docs |
| **i18n** | ✅ RU/EN/KK/ES (433 keys × 4 локали, 100% parity) | + insights namespace (19 keys × 4) | + developer namespace (44 × 4), changelog (12 × 4), pwa (12 × 4) | — |
| **Команда** | соло | +1 frontend (contractor or full-time) | +1 backend | +1 DevRel/marketing |
| **Метрики** | DAU, signups, activation | MRR (≥1), retention W2/W4 | MRR ≥ $1k, conversion ≥ 5% | MRR ≥ $5k, NRR ≥ 110% |

---

## Tech debt long list (горизонт 12+ мес)

В порядке убывания приоритета:

1. ✅ **Test coverage**: 0% → **42.5%** (Q1, цель 40% выполнена) → 60% (Q2) → 70% (Q3+)
2. **Observability**: ✅ stdout+histograms+Uptime Kuma (Q1) → Sentry + Grafana (Q4) → SLO dashboards (Q4)
3. **CI/CD**: ✅ reusable workflows + path-filters + SAST (Q1) → staging + preview deploys (Q4 — Dokploy access)
4. ✅ **DB migrations**: only up → up+down (Q1, golang-migrate с pgx/v5) → schema versioning (Q4)
5. **Secrets**: `.env` → Doppler/Vault (Q4)
6. **Error handling**: ad-hoc → typed errors + structured response (Q4 — RFC 7807 ProblemDetails)
7. ✅ **Frontend**: vanilla useState → react-query + react-router-dom + react-hook-form + zod + sonner (Q1-Q3, FSD structure complete)
8. **No event sourcing/audit trail** для team actions (Q4)
9. ✅ **ClickHouse perf**: not benchmarked → events_hourly_agg MV (Q2, 3-7× speedup, 35× storage reduction) → daily aggregate poверх hourly (Q4 при 30M+/day)
10. ✅ **Mobile**: PWA (Q3, installable + Web Push + adaptive layout vs native — chose PWA по cost/value tradeoff)
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

_Последнее обновление: 2026-05-10 (Q1+Q2+Q3 retro, mostly closed pre-quarter-start)._

---

## Q1+Q2+Q3 retro (2026-05-10, formally Q1 не начат — June 2026 +)

Все три квартала **closed ahead of schedule**. Suммарно:

- **Q1**: 9/11 done. Висят: E2E Playwright suite, marketplace distribution (external — Apple Dev/MS Partner accounts).
- **Q2**: 6/8 done. Sentry+penetration test отложены per design (Sentry — до paying, pentest — Q4 SOC2 prep).
- **Q3**: 8/11 done. Висят: Blog/Public roadmap/Case studies (content), CI/CD staging+Secrets vault (Dokploy/external setup).

### Сделано вне Q1-Q3 плана

- i18n × 4 локали (RU/EN/KK/ES) — планировался RU+EN. Сейчас 5+ namespaces × 4 = ~600 keys parity.
- Backend split refactor (был Q2) — выполнен в Q1.
- Frontend на react-query + react-router (был Q3) — выполнен в Q1.
- ClickUp вместо Linear/Jira — переориентирован per user feedback.
- GitLab support в PR-comment — bonus, ROADMAP только GitHub упоминал.
- **Caddy migration** (nginx → Caddy 2.11) — infra cleanup.
- **shadcn/Radix Select** + SimpleSelect helper.
- **Comprehensive mobile audit** — 15 файлов responsive pass, native-app feel в PWA.
- Tech-lead CI pipeline с reusable workflows + path-filters + SAST/secret/CVE scanners.
- `.trivyignore` с justification per CVE.

### Что висит (priority backlog)

1. **Tables → card-stack mobile** — последний UX-todo из mobile audit (commit/event/admin tables на phone).
2. **Typed errors backend** (RFC 7807 ProblemDetails) — Q4 tech-debt, можно начать раньше.
3. **E2E Playwright** — Q1 tech-debt-висяк, расстраховка для Q2-3 рефакторинга.
4. **Marketplace distribution для агентов** — нужны Apple Dev / MS Partner / Chrome Web Store аккаунты (user action).
5. **CI/CD staging environment** — Dokploy access нужен.
6. **Secrets vault setup** — Doppler / 1Password Connect аккаунт.

### Q4 backlog (March-May 2027)

В оригинальном плане:
- SSO (SAML/OIDC) — enterprise hard requirement >$500/mo
- Audit log — кто что когда
- SOC2 path (Vanta/Drata)
- Self-hosted enterprise edition
- Performance audit (CH partitioning, Redis cache)
- Disaster recovery (PG+CH backups, RTO<4h, RPO<1h)
- Multi-region readiness
