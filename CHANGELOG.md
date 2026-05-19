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
