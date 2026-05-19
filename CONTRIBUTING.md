# Contributing to Eye of Providence

Thanks for taking time to contribute. This guide gets you from a fresh clone
to a green PR in under 10 minutes.

## TL;DR

```bash
git clone https://github.com/luckyrogue/eye-of-providence.git
cd eye-of-providence
cp .env.example .env                       # adjust CHANGE_ME values
make dev                                   # docker compose up postgres + clickhouse + redis + dashboard
pnpm install --frozen-lockfile             # install JS deps (root + workspaces)
cd backend && go mod download && cd ..     # warm Go cache
cd agent/src-tauri && cargo fetch && cd ../..   # warm Rust cache
```

Open http://localhost:5173 for dashboard, http://localhost:8080 for backend API.

## Architecture overview

- `backend/` — Go monorepo, bounded-context layout (`internal/<bc>/`,
  `internal/<bc>app/`, `internal/<bc>/domain/`). Fiber for HTTP. See
  [`docs/architecture-bc.md`](docs/architecture-bc.md) and
  [`docs/data-model.md`](docs/data-model.md).
- `dashboard/` — React + TS + Vite, Feature-Sliced Design (FSD) layers
  (`shared/`, `entities/`, `features/`, `widgets/`, `pages/`, `app/`).
- `agent/` — Tauri 2 desktop agent. `src-tauri/` is Rust; `src/` is React UI.
- `ui/` — shared component library (`@eop/ui`), used by both dashboard and
  agent React frontends.
- `browser-extension/` — MV3 extension (Chrome / Edge / Brave).
- `ide-vscode/` — VS Code extension (compatible with Cursor, VSCodium).
- `infra/` — docker-compose for dev (`docker-compose.dev.yml`) and self-host
  prod (`docker-compose.full.yml`).
- `.github/workflows/` — CI: `ci.yml` orchestrates `_backend`, `_frontend`,
  `_agent-rust`, `_docker`, `_security`, `_e2e`, `_deploy`. `release.yml` fires
  on `v*.*.*` tag. `continuous-build.yml` fires on every push to `main`.

## Coding standards

| Stack | Formatter | Linter | Tests |
|---|---|---|---|
| Go | `gofmt -s` (auto) | `golangci-lint run ./...` | `go test ./...` |
| TS/TSX | `prettier --write` (lint-staged) | `eslint src` | `pnpm test` (vitest) |
| Rust | `cargo fmt` | `cargo clippy -- -D warnings` (non-blocking in CI for now) | `cargo test --lib` |

Pre-commit hook (`husky` + `lint-staged`) runs prettier on staged TS/JSON/MD.
You don't need to invoke it manually.

## Testing

- **Backend:** `cd backend && go test ./...` — unit. Integration tests need
  Postgres + ClickHouse (run `make dev` first), then `go test -tags=integration ./...`.
- **Frontend:** `pnpm test` (root) → runs vitest across workspaces.
- **Agent:** `cd agent/src-tauri && cargo test --lib --offline`.
- **E2E:** `pnpm e2e:test` (requires `pnpm e2e:install` once for browsers).

CI runs all of these on PR. If `_backend / integration` is red, look at the
artifact-uploaded logs — almost always a missing migration or a route
ordering issue (see `backend/internal/app/public_routes_test.go`).

## Commit format

Conventional Commits. Type prefixes:

| Type | Use |
|---|---|
| `feat` | new user-facing capability |
| `fix` | bug fix |
| `chore` | tooling, deps, infra |
| `docs` | documentation only |
| `refactor` | code restructure with no behavior change |
| `test` | adding/improving tests |
| `style` | formatting only (rare — usually auto via prettier/gofmt) |
| `ci` | workflows |
| `perf` | performance work |
| `revert` | revert of earlier commit (include `Reverts <sha>`) |

Scope is optional but useful: `fix(agent): parking_lot mutex`.

Body: explain the *why*, not the *what*. Include before/after when subtle.
Reference cluster IDs from [`docs/tech-debt.md`](docs/tech-debt.md) (C2, C13,
etc.) when relevant.

## Pull request review

- `CODEOWNERS` auto-requests `@luckyrogue` for security-critical paths
  (auth, plans, ingest, publicapi, audit, teams, webhooks, infra). Until
  team grows: solo-maintainer self-merge is OK for non-critical changes.
- Required CI checks (when branch protection is enabled — see
  [`docs/ci-hardening.md`](docs/ci-hardening.md)): backend lint/unit/integration,
  agent-rust cargo check, all security scans, docker build + sign, e2e.

## Releases

```bash
git tag -a v0.1.0-alpha.3 -m "Release notes here..."
git push origin v0.1.0-alpha.3
```

`release.yml` then:
1. Builds Tauri agents for 4 targets (macOS aarch64/x86_64, Windows, Linux).
2. Packages VS Code `.vsix` and browser ext `.zip`.
3. Creates GitHub draft Release with all 11 artifacts.
4. Maintainer publishes manually after smoke-test.

Code-signing secrets (`APPLE_*`, `WINDOWS_*`) are optional; without them
binaries are unsigned (users see Gatekeeper/SmartScreen warning on first
launch).

Update [`CHANGELOG.md`](CHANGELOG.md) before tagging: move `[Unreleased]`
entries into the new version section.

## Security disclosures

Do **not** open public GitHub issues for vulnerabilities. Email
`d.rysdovletov@gmail.com`. See [`.github/SECURITY.md`](.github/SECURITY.md)
for the full policy and image-verification recipes.

## Where to start

- Browse [`docs/tech-debt.md`](docs/tech-debt.md) — 13 clusters, prioritized.
  Items labeled S0/S1 are good first contributions for serious engagement.
- Issues tagged `good-first-issue` (when we open them up).
- Docs improvements always welcome — `docs/` has a few stubs that need flesh.

## Code of conduct

Be kind. Disagree on technical merits; don't punch down. This is a small
project; harassment policy is "don't, or you're out."
