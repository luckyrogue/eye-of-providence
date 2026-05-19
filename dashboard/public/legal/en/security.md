# Security — Eye of Providence

**Last updated:** 2026-05-19 · **Version:** v0.1 alpha

This page describes how we handle security: what's in place today, how to
report a vulnerability, and what to expect in response. For the internal
engineering threat model see
[`docs/threat-model.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/threat-model.md);
this page is the user-facing summary.

## Reporting a vulnerability

**Please do not open a public GitHub issue.** Email
**`main@rysdavletov.org`** with:

- A description of the issue and its impact
- Reproduction steps or a proof-of-concept
- Affected version or commit SHA
- Your contact for follow-up

We aim to **acknowledge within 48 hours** and provide a remediation
timeline within **5 business days**.

In-scope: backend API, dashboard, agent (desktop / browser ext /
VS Code), Docker image and CI infrastructure.

Out of scope: self-hosted instances run by third parties, third-party
services (ClickHouse Cloud, Resend, Dokploy), DoS testing (we have
rate-limiting but it's not a vulnerability surface to exploit).

## Supported versions

Security fixes land on `main`. During alpha we do not maintain LTS
branches.

| Version              | Supported |
| -------------------- | --------- |
| `main`               | ✅        |
| `v0.1.x-alpha.*`     | ✅ (latest only) |
| pre-alpha rolling tags | ❌      |

## Current security posture

Concrete controls in place today:

### Backend & data

- **Auth:** bcrypt password hashing (cost 10), JWT HS256 with
  `token_version` revocation. WebAuthn / passkey support for
  second-factor.
- **Rate limits:** 10 req/min on auth endpoints, 120 req/min on `/v1/*`.
- **Per-user isolation:** every analytics query filters by JWT subject;
  cross-user access is impossible at the SQL layer.
- **GDPR:** `GET /v1/me/export` returns all your data as JSON;
  `DELETE /v1/me/data` permanently wipes it.

### Image and supply chain

- **Cosign signed:** every Docker image pushed to `ghcr.io/luckyrogue/eop`
  is signed with Sigstore keyless OIDC. Verify with:
  ```bash
  cosign verify ghcr.io/luckyrogue/eop:<sha> \
    --certificate-identity-regexp '^https://github.com/luckyrogue/eye-of-providence/' \
    --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
  ```
- **SLSA Build L3 provenance:** verifiable proof the image was built by
  our CI from a specific commit. Verify with `gh attestation verify
  oci://ghcr.io/luckyrogue/eop:<sha> --owner luckyrogue --predicate-type
  'https://slsa.dev/provenance/v1'`.
- **CycloneDX SBOM:** every image ships a Software Bill of Materials as
  an attestation. Fetch with `gh attestation download`.

Full verification recipes:
[`.github/SECURITY.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/.github/SECURITY.md).

### Agent (desktop)

- **Encrypted local buffer:** events stored in SQLite are encrypted
  with AES-256-GCM. The key lives in the OS keyring (macOS Keychain,
  Windows Credential Manager, GNOME keyring).
- **Pairing tokens** stay in the keyring, never written in plaintext.
- **Privacy invariants:** the agent never reads file contents, prompts,
  AI replies, raw keystrokes, or clipboard text. Only counts, hashes,
  and timestamps leave the device. See
  [Privacy Notice](/privacy) §1 for the full data map.

### CI / development

- CodeQL static analysis (Go + JS/TS) on every PR.
- Trivy + OSV scans on source and image; dependency-review on PR.
- gitleaks scans every commit for accidentally committed secrets.
- Step-Security `harden-runner` audits runner egress.

### Known limitations (transparent)

- **Alpha installers are unsigned.** Apple Developer ID and Windows EV
  cert are not yet purchased. See
  [Why is this unsigned?](/docs/install#почему-installer-не-подписан)
  in the install guide for workarounds. Image signing (Cosign) is
  unaffected — the backend image is fully verifiable.
- **Branch protection** is being enabled per
  [`docs/ci-hardening.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/ci-hardening.md)
  during alpha-1 follow-up. Until then, all merges go through CI but
  not enforced reviewers.
- **PITR for Postgres** not configured; RPO = 24 h (daily dump). Tighter
  RPO targeted for GA. Details in
  [`docs/disaster-recovery.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/docs/disaster-recovery.md).

## Responsible disclosure

We do not currently offer a bug bounty. We do credit researchers (with
permission) in the relevant CHANGELOG entry and the commit message of
the fix.

Please give us a reasonable window — typically the timeline we agree on
in the initial acknowledgement, default 90 days — before public
disclosure.

## Updates to this page

Material changes to security posture are tracked in
[`CHANGELOG.md`](https://github.com/luckyrogue/eye-of-providence/blob/main/CHANGELOG.md)
under the **Security** subsection of each release.
