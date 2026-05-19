# Tech Debt Register

**Created:** 2026-05-19, post `v0.1.0-alpha.2` ship.
**Method:** 5 parallel domain audits (backend / frontend / agent / infra-CI /
docs-security-deps) by automated explorers + manual review.
**Snapshot count:** ~95 individual findings, normalized into the clusters below.

This doc is the **execution backlog** for closing tech debt before promoting
alpha → beta. Items are grouped by **remediation cluster** (so they batch
into one PR/sprint), not by severity — but each item carries a severity tag
so prioritization across clusters is possible.

## Severity legend

| Tag | Meaning |
|---|---|
| **S0** | Crash, data loss, security exposure, blocks production-grade SLA |
| **S1** | Architecture debt, fragile-by-design, ops risk, blocks beta promotion |
| **S2** | Polish, coverage gap, dev-ergonomics, can ship without |

A cluster is rated by its **highest-severity item**.

---

## Cluster summary

| # | Cluster | Severity | Items | Est. effort |
|---|---|---|---|---|
| C1 | Supply chain — code-signing + updater | **S0** | 4 | 1 week |
| C2 | Agent runtime stability — mutex/CSPRNG | **S0** | 3 | 2 days |
| C3 | CI hardening — branch protection + envs | **S0** | 5 | UI + 1 day |
| C4 | Router middleware contract | **S1** | 2 | 1 day |
| C5 | Clean-architecture refactor finish | **S1** | 6 | 1 week |
| C6 | Cross-platform parity (agent) | **S1** | 2 | 1 sprint |
| C7 | Backend test coverage push | **S1** | 7 | 1 sprint, ongoing |
| C8 | Docs — terminology, structure, freshness | **S1** | 10 | 2 days |
| C9 | Phase-blocked code (Apple OAuth, SSO, Phase-3 admin) | **S1** | 3 | scope-dependent |
| C10 | Frontend i18n + architecture polish | **S2** | 8 | 3 days |
| C11 | Hardcoded constants → config | **S2** | 6 | 1 day |
| C12 | Observability + DR readiness | **S2** | 4 | 1 week |
| C13 | Dependency drift + small ops fixes | **S2** | 6 | continuous |

Total **S0** = 12 items across 3 clusters · **S1** = ~30 across 6 · **S2** = ~50.

---

## C1 — Supply chain: code-signing + updater [S0]

Currently shipping unsigned binaries to public (alpha-2). Gatekeeper /
SmartScreen warning on every install. No auto-update path.

| Item | Sev | Where | Detail |
|---|---|---|---|
| Apple notarization absent | S0 | `release.yml:96` | `tauri-action@v0` does NOT invoke `notarytool`. Need v0.5+ OR custom step with `APPLE_ID`/`APPLE_PASSWORD`/`APPLE_TEAM_ID`. |
| Windows EV signing absent | S0 | `release.yml` (no `signtool`) | `WINDOWS_CERTIFICATE` env exported but never used by build step. |
| Tauri updater not configured | S0 | `agent/src-tauri/tauri.conf.json` | No `updater` block. Requires signed manifest → requires notarization first. |
| macOS universal binary not merged | S2 | `release.yml:49-52` | aarch64 + x86_64 built as separate artifacts. No `lipo -create` merge. |

**Blockers:** Apple Developer account + cert ($99/yr), Windows EV cert
(~$300/yr through Sectigo/DigiCert). Until then alpha stays unsigned.

---

## C2 — Agent runtime stability: mutex + CSPRNG [S0]

| Item | Sev | Where | Detail |
|---|---|---|---|
| Mutex poisoning in store | S0 | `agent/src-tauri/src/core/store.rs:87,106,173,190,206,216` | 6× `.lock().unwrap()`. If any thread panics holding lock, ALL subsequent .unwrap() panic the whole agent. Fix: `parking_lot::Mutex` (never poisons). |
| Weak CSPRNG for local API token | S0 | `agent/src-tauri/src/lib.rs:489-496` | `rand_byte()` uses `SystemTime::subsec_nanos()` as entropy. Predictable if clock known. Should use `rand::rngs::OsRng` (`crypto.rs` already does this correctly). |
| Silently dropped idle events | S1 | `agent/src-tauri/src/core/watcher.rs:120` | `let _ = store.push(&idle_event)`. Idle events lost on push error, no log. |

---

## C3 — CI hardening: branch protection + envs [S0]

| Item | Sev | Where | Detail |
|---|---|---|---|
| Branch protection NOT set on `main` | S0 | GitHub UI | No signed-commits requirement, no required reviewers, no required status checks, force-push allowed. |
| GitHub Environment `production` no approval | S1 | `_deploy.yml:20`, GitHub UI | Env declared but no manual approver + 5min wait timer. Typo in DOKPLOY_WEBHOOK ships to prod. |
| `DOKPLOY_WEBHOOK` undocumented | S1 | `.env.example`, `_deploy.yml:194` | Used in workflow but not in env example. |
| Conventional-commits PR check missing | S1 | new workflow needed | Blocks auto-changelog + semver bumping. |
| Harden-runner audit-only | S2 | 9 workflows × `egress-policy: audit` | Planned flip to `block` after baseline; explicit allow-list missing for GHCR/npm/GitHub. |

---

## C4 — Router middleware contract [S1]

We hit the same Fiber `app.Group("/v1", mw)` ≡ `app.Use("/v1", mw)` trap
twice (password-reset and devices/pair). Pattern is fragile.

| Item | Sev | Where | Detail |
|---|---|---|---|
| Public-routes registration order is implicit | S1 | `backend/internal/app/modules.go:60-78` | Comment marks the trap, but no compile-time guarantee. Next public route added by anyone unaware will 401. |
| 15 modules export independent `RegisterRoutes` | S1 | various `internal/*/handler.go` | No central registry. One missed registration silently drops endpoint. |

**Fix shape:** introduce `RouteRegistry` with explicit `Public()` /
`Authenticated()` buckets, OR move auth check to per-route middleware
(slower but explicit), OR add unit test that pings each known public
endpoint expecting 200/400 not 401.

---

## C5 — Clean-architecture refactor finish [S1]

Recent commits split `internal/X/` → `X/`, `Xapp/`, `domain/`, `X_adapters.go`.
Some leftovers and inconsistencies.

| Item | Sev | Where | Detail |
|---|---|---|---|
| `internal/auth/` no `domain/` package | S1 | `backend/internal/auth/` | Unlike 13 other modules. `User`, `Claims`, `Identity` live in top-level + handler files. |
| Duplicate cache impl | S1 | `backend/internal/cache/` + `internal/content/cache.go` | Near-duplicate with same `DefaultCacheTTL = 5min`. |
| Bare SQL in handler layer | S1 | `backend/internal/teams/handler.go:88,178` | `SELECT email FROM users` directly in HTTP handler, silently ignored errors. |
| Adapters scattered | S1 | `backend/internal/teams/` | `adminapp_adapters.go`, `membersapp_adapters.go`, etc — no single composition point. |
| Phase 3 stub file | S2 | `backend/internal/teams/admin_phase3.go` | 10 unimplemented funcs, unreferenced from handlers/adapters. |
| Migration 023 missing | S2 | `backend/internal/migrate/sql/postgres/` | 001-022, 024-025 present. 023 gap. Suggests partial rollback chain or undocumented skip. |

---

## C6 — Cross-platform parity (agent) [S1]

| Item | Sev | Where | Detail |
|---|---|---|---|
| Windows no keystroke/clipboard tracking | S1 | `agent/src-tauri/src/platform/windows/mod.rs` | Inherits trait defaults returning 0/None. Windows users get analytics gap vs macOS. |
| Linux silently `NoopWatcher` | S1 | `agent/src-tauri/src/platform/mod.rs:59-75` | No Linux module at all. Linux installer ships but agent collects nothing. |

**Decision needed:** is Linux support a v0.x goal, or v1.0+? Either build
the module or add startup warning + landing copy disclosure.

---

## C7 — Backend test coverage push [S1]

Critical packages under 30% coverage. Mutations land untested.

| Package | Coverage | Sev |
|---|---|---|
| `internal/auth` | **9.4%** | S1 |
| `internal/auth/sessionapp` | **0%** | S1 |
| `internal/auth/passwordresetapp` | **0%** | S1 |
| `internal/teams` | **1.8%** | S1 |
| `internal/teams/adminapp` | **6.1%** | S1 |
| `internal/teams/registrationapp` | **0%** | S1 |
| `internal/store` (CH integration) | **6.1%** | S1 |
| `internal/attribution/attributionapp` | **0%** | S1 |
| `internal/webhooks/webhooksapp` | **5.7%** | S1 |

Coverage gate currently `default: 12%` in `_backend.yml:21`. Spec says
ratchet -2pp per sprint but no enforcement. Should set target by package.

---

## C8 — Docs: terminology, structure, freshness [S1]

| Item | Sev | Where | Detail |
|---|---|---|---|
| 4 docs still say "closed beta" | S1 | `docs/beta-install.md`, `ide-vscode/README.md:45`, `AGENTS.md:1`, `docs/disaster-recovery.md:1` | Project canonical is now "early access" / "alpha". |
| No `CHANGELOG.md` | S1 | repo root | Only GitHub Releases (was draft until publish). No machine-readable change log. |
| No `CONTRIBUTING.md` | S1 | repo root | New-dev onboarding absent (architecture, code review guidelines, design principles). |
| No API reference | S1 | `docs/api/` empty dir | REST endpoints undocumented outside Go code. |
| Threat model stale | S1 | `docs/threat-model.md` (last 2026-05-04) | Missing passkey/WebAuthn STRIDE, GDPR export, admin panel, Resend/Dokploy. |
| `architecture-bc.md` stub | S1 | `docs/architecture-bc.md` (57 lines) | Just DDD checklist, no actual architecture rationale. |
| Orphan docs not linked from README | S1 | `docs/agents-publishing.md`, `integrations-pr-comment.md`, `disaster-recovery.md` | Discoverability=0 for non-contributors. |
| README missing quick-start + version | S1 | `README.md` | No "get started in 5 min". No `v0.1.0-alpha` indicator. No download links. |
| Self-host docs reference wrong service name | S2 | `docs/self-hosting.md:27`, `disaster-recovery.md:182` | `eop-app` vs `eop` mismatch (compose has `eop`). |
| Quarterly DR test not scheduled | S2 | `docs/dr-test-log.md` referenced but doesn't exist | Per `disaster-recovery.md:201-210` mandate. |

---

## C9 — Phase-blocked code [S1]

Code paths that explicitly reject inputs with "Phase N" rationale. Some
are wired into prod, will return errors to real users if hit.

| Item | Sev | Where | Detail |
|---|---|---|---|
| Apple OAuth blocked at Phase 2 boundary | S1 | `backend/internal/auth/apple_test.go:59` + production code | Apple Exchange "must NOT mint identities until Phase 2". |
| Non-OIDC SSO rejected at runtime | S1 | `backend/internal/sso/registry.go:52` + `service.go` | Live rejection, not test-only. |
| Plan overrides pending Phase 3 wiring | S1 | `backend/internal/teams/admin_plan_overrides_test.go:141` | Admin endpoint registered but incomplete. |

---

## C10 — Frontend i18n + architecture polish [S2]

| Item | Sev | Where | Detail |
|---|---|---|---|
| Spanish untranslated strings | S2 | `dashboard/src/shared/i18n/locales/es/common.json:10,17,23,33,36,40` | "Admin", "Insights", "Workspace", "alpha", etc. left English. |
| `kk/common.json` mixed alphabet | S2 | (already partially fixed) | Sweep again post-alpha rename. |
| Feature module structure inconsistent | S1 | `dashboard/src/features/*` | Some have `api/` + `ui/`, others flat. 15+ features without `api/` folder despite having data fetching. |
| 127 inline `style={{}}` blocks | S2 | dashboard-wide | HSL CSS-var refs that should be Tailwind utilities or CSS module. |
| TODO: default providers fallback | S1 | `dashboard/src/entities/user/api/{req.ts:92,types.ts:43}` | Remove once `/v1/auth/config` ships. |
| Stub admin overview stats | S2 | `dashboard/src/widgets/admin-overview/ui/overview.tsx:3-5` | Comments "TODO Total teams (с beta limit indicator)" — unimplemented. |
| Dead `GithubIcon` | S2 | `dashboard/src/pages/landing/ui/hero.tsx` | Defined, never used. |
| Mixed cache invalidation patterns | S2 | features `onSuccess` callbacks | No centralized invalidation strategy. Brittle when features cross-touch. |

---

## C11 — Hardcoded constants → config [S2]

| Item | Sev | Where | Detail |
|---|---|---|---|
| JWT TTL hardcoded 14d | S2 | `backend/internal/auth/handler.go:15` | `const tokenTTL = 14 * 24 * time.Hour`. |
| Handoff session TTL 30s | S2 | `backend/internal/auth/handler.go:18` | Affects password reset / device pair callbacks. |
| Reset link expiry 1h | S2 | `backend/internal/auth/passwordresetapp/service.go:10` | `defaultResetTTL`. |
| WebAuthn session 5min | S2 | `backend/internal/auth/webauthn.go:26` | `webauthnSessionTTL`. |
| Ingest batch limit 5000 | S2 | `backend/internal/ingest/handler.go:25` | `maxEventsPerBatch`. |
| Bcrypt cost 10 | S2 | `backend/internal/auth/password.go:23` | `bcrypt.DefaultCost`. Should be tunable per deploy. |
| Agent GC retention 7d / interval 1h | S2 | `agent/src-tauri/src/lib.rs:250` | Compile-time constants. |

---

## C12 — Observability + DR readiness [S2]

| Item | Sev | Where | Detail |
|---|---|---|---|
| No Prometheus / Grafana setup | S2 | no `infra/observability/` | Backend has `/metrics`, but no scraper / dashboard. Dokploy only does container-level health. |
| Alert rules absent | S2 | — | No Alertmanager config. No queries on latency / error rate / data loss. |
| PITR not configured | S1 | `docs/disaster-recovery.md:107-111` | Acknowledged as deferred. 24h RPO ceiling. Blocks payment-grade SLA. |
| Canary deploy + rollback | S2 | `_deploy.yml` | Currently webhook → healthz → done. No staging promotion. No `:staging` → `:latest`. |

---

## C13 — Dependency drift + small ops [S2]

| Item | Sev | Where | Detail |
|---|---|---|---|
| pnpm version inconsistency | S1 | `_e2e.yml:39` pins 9.12.3; others don't | Lockfile drift if `pnpm/action-setup@v6` auto-bumps. |
| Node 20 vs 22 mismatch | S2 | workflows say 20, `Dockerfile:42` says 22 | Pick one. |
| Alpine pin mismatch | S1 | `dev.yml` floats `postgres:16-alpine`, `full.yml` pins `16.6-alpine` | Dev silent upgrades vs prod stable. |
| `golang.org/x/crypto` lag | S2 | `backend/go.mod` | 0.51.0, latest ~0.55. Not critical. |
| Browser ext hardcoded host | S1 | `browser-extension/manifest.json:5` | `eop.rysdavletov.org/api/*` baked into MV3 manifest. Domain change requires rebuild + reinstall. |
| Hardcoded download URL in PR-comment formatter | S1 | `backend/internal/prcomment/formatter.go:12` | `https://eop.rysdavletov.org/downloads`. |

---

## Suggested phasing

### Sprint 1 (week 1) — “alpha → beta gate”
- **C2** Agent runtime stability (mutex / CSPRNG / idle event log) — 2 days
- **C3** Branch protection + production env approval — UI + 1 day code
- **C4** Router middleware contract — 1 day (per-endpoint test)
- **C8** Docs terminology sweep + CHANGELOG.md + CONTRIBUTING.md — 2 days
- **C13** Pin pnpm/Node/Alpine + dep drift — 0.5 day

### Sprint 2 (week 2) — “stop the bleeding”
- **C1** Code-signing (Apple cert + Windows EV) — once certs purchased
- **C5** Finish clean-arch refactor: auth/domain, dedupe cache, kill bare SQL — 1 week
- **C7** Backend coverage push to 35% target (focus auth, teams) — ongoing

### Sprint 3 — “quality + parity”
- **C6** Windows keystroke/clipboard parity (or document macOS-only) — 1 sprint
- **C10** Frontend i18n sweep + feature-module shape enforcement — 3 days
- **C12** Prometheus/Grafana + canary deploy — 1 week
- **C11** Externalize 6 hardcoded constants → env vars — 1 day

### Continuous / out-of-band
- **C9** Phase-blocked code unblocks when business is ready (Apple OAuth, SSO non-OIDC, Phase-3 admin)

---

## How to use this doc

- When you start work on a cluster, open a tracking issue titled `[tech-debt] CN — <cluster name>` and link this section.
- When closing items, edit this file with `~~strikethrough~~` and a commit ref.
- Re-audit after each minor release (`v0.2.0`, `v0.3.0`) — append new clusters, don't rewrite existing.
- The cluster IDs are stable references; severities can change but item count per cluster grows/shrinks.

## Maintenance

This document is generated by audit, then hand-curated. To re-run the
audit pass:
```
# Each domain in its own focused agent (~2-3 min wall-clock):
# - backend Go
# - frontend TS/React
# - agent Rust/Tauri
# - infra/CI/Ops
# - cross-cutting docs/sec/deps
```
Then synthesize into clusters with severity ratings. Don't drop a cluster
just because it's empty next time — mark it `✅ resolved (commit ref)`.
