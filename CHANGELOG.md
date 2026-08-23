# Changelog

All notable changes to Eye of Providence are documented in this file.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Pre-release identifier `-alpha.N` is used for early-access builds before v1.0.

## [Unreleased]

### Added
- Backend test that fails CI if a public `/v1/*` endpoint is mistakenly placed
  behind `auth.Middleware` due to Fiber `app.Group("/v1", mw)` ≡ `app.Use`
  semantics. `backend/internal/app/public_routes_test.go`.
- Tech-debt register `docs/tech-debt.md` — 13 clusters covering ~95 findings
  from post-alpha audit.
- **Attribution worker now ships and runs.** Built as a third binary into the
  app image and supervised by the container entrypoint alongside `api` and
  `caddy`; toggle with `EOP_WORKER_ENABLED`. Previously `cmd/worker` was never
  built into the image nor declared in any compose file, so
  `attribution_events` was never populated in any deployed environment.
- `GET /v1/summary/provenance` — code-provenance aggregate (`focus_ms` per
  category) over `attribution_events`. Documented in `docs/api/openapi.yaml`.
- Release-readiness snapshot `docs/internal/release-readiness.md` — evidence-backed
  audit of what does and doesn't reach users.
- Tests: `classify()` Phase A matrix (`internal/attribution` was at 0%),
  `AggregateProvenance` across memory/cached stores, analytics service, and an
  e2e contract check on the new endpoint.

### Changed
- Agent local SQLite store now uses `parking_lot::Mutex` (never poisons) instead
  of `std::sync::Mutex` — eliminates cascade-panic risk in watcher/tray/ingest
  loops.
- Local-API token generator switched from `SystemTime::subsec_nanos` to
  `rand::rngs::OsRng` (cryptographically secure).
- CI workflows pin pnpm `9.12.3` everywhere; Node bumped 20 → 22 to align with
  `Dockerfile`. `infra/docker-compose.dev.yml` patch-pinned to match prod
  (`postgres:16.6-alpine`, `clickhouse:24.10.4-alpine`, `redis:7.4-alpine`).
- Documentation: `docs/beta-install.md` renamed to `docs/alpha-install.md` with
  a 1-line redirect stub. Terminology sweep across `AGENTS.md`,
  `disaster-recovery.md`, `.github/SECURITY.md`.

### Fixed
- Watcher idle event push now logs failures via `tracing::warn` instead of
  silent drop.
- **Code-provenance donut now reads real attribution.** It was querying
  `/v1/summary/categories` (raw event categories) and papering over the
  taxonomy mismatch with two fallbacks, so three of its five labelled segments
  were structurally always zero. It now reads `/v1/summary/provenance`, its
  buckets match the `attribution_events` schema, and empty categories are not
  drawn at all — a `0%` segment reads as a measurement rather than the absence
  of one. Added an empty state for when no attribution exists yet.
- AI share in the donut centre counted only keys prefixed `ai_`, silently
  excluding `pasted_ai`. Now an explicit key set.
- Container entrypoint: under `set -e` a bare `wait -n` returning non-zero
  aborted the script before cleanup, orphaning the sibling processes.
  The worker runs in a 15s retry loop and is deliberately excluded from the
  liveness `wait -n`: it exits fatally when ClickHouse is unreachable, and in
  production ClickHouse is a separate stack — otherwise a CH blip would take
  down the API and dashboard with it. Analytics may lag; the UI may not.
- `.prettierignore` now covers local Playwright output (`e2e/playwright-report/`,
  `e2e/test-results/`), which made `pnpm format:check` fail for anyone who had
  run e2e locally.

### Documentation
- `README.md` §1.4 now states which provenance categories actually work today
  and what each requires — agent-only installs get `unknown`, and
  `pasted-AI`/`pasted-other` await Phase B.
- `docs/self-hosting.md` documents the attribution worker, `EOP_WORKER_ENABLED`,
  and the single-replica constraint (worker position in `worker_state` has no
  leader election; two replicas double-write).

## [0.1.0-alpha.2] - 2026-05-18

### Added
- Cross-platform agent installers under all four targets:
  - macOS Apple Silicon (`aarch64.dmg`), Intel (`x64.dmg`)
  - Windows MSI (`x64_en-US.msi`) + NSIS (`x64-setup.exe`)
  - Linux `.deb`, `.rpm`, `.AppImage`
- VS Code extension `.vsix` and browser extension `.zip` artifacts.
- Public alpha release published as GitHub pre-release.

### Changed
- Backend route registration order: public routes (`/v1/auth/forgot-password`,
  `/v1/auth/reset-password`, `/v1/devices/{pair,poll}`) now registered BEFORE
  `teams.RegisterRoutes` so they aren't caught by `/v1` `auth.Middleware`.

### Fixed
- `release.yml`: signing secrets exported to env only when non-empty
  (tauri-action treated empty-string Apple cert as set and failed import).
- `release.yml`: added `GITHUB_TOKEN` env to tauri-action so artifact upload
  to GH Release succeeds.
- Dashboard admin panel: content + email-templates routes use path params for
  locale (`/v1/admin/content/<slug>/<locale>`) matching backend expectation —
  previous query-param form returned 404 "Cannot GET".
- Agent pairing wizard links user to `/integrations` (where `DevicesWidget`
  lives); `DevicesWidget` also dual-published on `/settings` for users on
  already-shipped agent builds.
- Backend integration tests for `TeamRow` JSON serialization (added json
  tags) and `password_reset_test` (inlined moved helpers).
- macOS Linux build of agent: `permissions_status()` no longer references
  `platform::macos::accessibility_app_label()` from non-macOS targets.
- Various staticcheck issues (S1016 struct→cast, unused functions) in the WIP
  clean-architecture refactor.

### Security
- Cosign-signed GHCR image: `ghcr.io/luckyrogue/eop:bc45182`. Verifiable via
  Sigstore + SLSA Build L3 provenance (see `.github/SECURITY.md`).

## [0.1.0-alpha.1] - 2026-05-18 (never published)

Failed Release workflow due to release.yml bugs; superseded by alpha.2. Tag
remains for git-history continuity. Draft GH Release was deleted.

### Added
- First public alpha tag. Same source content as alpha.2 (which has the build
  pipeline fixes).

## Prior history

Before v0.1.0-alpha.1, development was on rolling `continuous-main` pre-release
tag (CI builds from every push to `main`). Full git history available via
`git log` — milestones surfaced in subsequent entries as we mature the
release cadence.

[Unreleased]: https://github.com/luckyrogue/eye-of-providence/compare/v0.1.0-alpha.2...HEAD
[0.1.0-alpha.2]: https://github.com/luckyrogue/eye-of-providence/releases/tag/v0.1.0-alpha.2
[0.1.0-alpha.1]: https://github.com/luckyrogue/eye-of-providence/releases/tag/v0.1.0-alpha.1
