# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Eye of Providence, **please do not open a public GitHub issue**.

Instead, email **d.rysdovletov@gmail.com** with:

- Description of the issue + impact
- Reproduction steps (or PoC)
- Affected version / commit SHA
- Your contact for follow-up

We aim to acknowledge within **48 hours** and provide a remediation timeline within **5 business days**.

## Supported Versions

Security fixes land on `main`. We do not maintain LTS branches during beta.

| Version    | Supported |
| ---------- | --------- |
| `main`     | ✅        |
| pre-beta tags | ❌     |

## Scope

In scope:
- Backend API (`backend/`)
- Dashboard (`dashboard/`, `ui/`)
- Agents (`agent/src-tauri`, `browser-extension/`, `ide-vscode/`)
- Docker image, CI/CD, deploy infrastructure (`infra/`, `Dockerfile`, `.github/`)

Out of scope:
- Self-hosted instances run by third parties (we only secure code; ops is on you)
- Third-party services (ClickHouse Cloud, Resend, Dokploy) — report to them directly
- DoS (we have rate-limiting; ddos protection is via deploy infra, not in our codebase)

## Hardening

Current security posture:
- **CI**: CodeQL (security-and-quality on PR, security-extended nightly), gitleaks, Trivy fs + image scans, SBOM
- **Image**: distroless-style nginx + static Go binary, non-root user
- **Migrations**: idempotent + advisory_lock; down-scenarios reviewed
- **Auth**: bcrypt (cost 10), JWT HS256 with token_version revocation, 1h password reset TTL
- **Secrets**: never committed; production via Dokploy env vars
- **Dependencies**: Dependabot weekly, pinned (lockfiles + go.sum verification)

See `infra/PRODUCTION.md` for runbook.
