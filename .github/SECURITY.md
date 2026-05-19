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

Security fixes land on `main`. We do not maintain LTS branches during alpha/beta.

| Version    | Supported |
| ---------- | --------- |
| `main`     | ✅        |
| `v0.1.x-alpha.*` | ✅ (latest only) |
| pre-alpha tags | ❌     |

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
- **CI**: CodeQL (security-and-quality on PR, security-extended nightly), gitleaks, Trivy fs+image, OSV-Scanner, GitHub dependency-review (PR-only), step-security/harden-runner egress audit
- **Image**: Alpine 3.23 + custom-built caddy v2.11.2 + static Go binary, non-root user
- **Supply chain**:
  - Cosign keyless signing (Sigstore OIDC, no key management)
  - SLSA Build L3 provenance attestation (`actions/attest-build-provenance`)
  - CycloneDX SBOM generated per build + pushed as registry attestation
- **Migrations**: idempotent + advisory_lock; down-scenarios reviewed
- **Auth**: bcrypt (cost 10), JWT HS256 with token_version revocation, 1h password reset TTL
- **Secrets**: never committed; production via Dokploy env vars
- **Dependencies**: Dependabot weekly grouped PRs; lockfiles enforced; license deny-list (GPL/AGPL) on PR

## Verifying release artifacts

Every image pushed to `ghcr.io/luckyrogue/eop:<sha>` is signed and attested. To
verify before deploying:

```bash
SHA=<commit-sha>                          # e.g. effbcf5...
IMG="ghcr.io/luckyrogue/eop:${SHA}"

# 1. Verify Sigstore signature — must come from this repo's workflows
cosign verify "$IMG" \
  --certificate-identity-regexp '^https://github\.com/luckyrogue/eye-of-providence/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'

# 2. Verify SLSA L3 build provenance — fails if image wasn't built by our CI
gh attestation verify oci://"$IMG" \
  --owner luckyrogue \
  --predicate-type 'https://slsa.dev/provenance/v1'

# 3. Fetch + inspect SBOM (CycloneDX JSON)
gh attestation download oci://"$IMG" --predicate-type cyclonedx
jq '.predicate.components[] | select(.name == "github.com/caddyserver/caddy/v2")' attestation.jsonl
```

See `infra/PRODUCTION.md` for runbook.
