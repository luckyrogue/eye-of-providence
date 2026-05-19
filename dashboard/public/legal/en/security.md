# Threat model — STRIDE

> **Status:** ⚠ last reviewed 2026-05-04 — re-review required before alpha → beta
> promotion. Known coverage gaps (added after last review): passkey/WebAuthn
> auth, GDPR-export endpoint (`GET /v1/me/export`), admin panel (super-admin
> view of aggregates + audit log), third-party integrations (Resend email,
> Dokploy hosting). Tracked in [`tech-debt.md`](tech-debt.md) C8.

Scope: backend (`cmd/api`), desktop agent (Tauri), browser extension (MV3), VS Code plugin, Claude Code hooks.

## STRIDE by component

### Backend (Go)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Forged events with another user's user_id | JWT rebinds event.user_id from the token in `ingestapp/service.go` (`e.UserID = userID`); client-supplied user_id is ignored |
| **T**ampering | Event tampering in transit | HTTPS in production (see `docs/self-hosting.md` production checklist) |
| **R**epudiation | "I did not send that" | `eop_ingest_events_*` metrics + access logs (Fiber middleware/logger) |
| **I**nformation disclosure | Leak of another user's data via analytics | All analytics endpoints filter by `claims.UserID`; cross-user queries are impossible |
| **D**oS | Event flooding | Fiber limiter 120 req/min on `/v1/*` (`cmd/api/main.go`); `domain.ValidEvent` drops durations >24h; dedicated per-ingest Redis limiter — roadmap |
| **E**levation of privilege | dev-token in production | `EOP_ENABLE_DEV_TOKEN` forbidden when `EOP_ENV=production`; route returns 404 when disabled (`config.go`, `dev_token_test.go`) |

### Desktop agent (Tauri)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Foreign process sends events via local API | Bearer token in `~/<data>/eop.local-token`, verified in `core/local_api.rs:handle` |
| **T**ampering | SQLite buffer modification | Local file with user-only permissions; AES-256-GCM at rest (`agent/src-tauri/src/core/crypto.rs`, `store.rs`) |
| **R**epudiation | — | Local-only, no multi-user |
| **I**nformation disclosure | Collection of file content / prompts | **Architectural invariant**: agent does not read Claude Code hook stdin, does not parse file bodies; only timestamps, char counts, hashes. Violation = bug |
| **D**oS | Disk fill via event_buffer | TTL and batch flush; in Phase 8 — hard limit on pending_count |
| **E**levation of privilege | macOS Accessibility permission | Requested explicitly via onboarding flow; without it keystroke counts are unavailable (graceful degradation) |

### Browser extension (MV3)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | DOM content-script spoofing | content script reads only selection size + host; content is not serialized |
| **T**ampering | XSS on page → fake event injection | Events go through `chrome.runtime.sendMessage` — sender verified by service worker (host whitelist) |
| **I**nformation disclosure | Accidental URL/title/content upload | `host_permissions` whitelisted (AI domains + localhost only); content scripts send ONLY `host` + `size` |
| **E**levation of privilege | OAuth cookie theft via extension | Extension has no access to foreign cookie stores; JWT in `chrome.storage.local` (isolated per extension) |

### VS Code extension

| Threat | Vector | Mitigation |
|---|---|---|
| **I**nformation disclosure | File content via diff | `onDidChangeTextDocument` gives lengths and timestamps; actual text does not enter our payload (see `extension.ts::onChange`) |
| **T**ampering | Token in settings.json — anyone with file access | Planned replacement: `secrets.SecretStorage` API (V1) |

### WebAuthn / passkeys

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Credential replay | Challenge stored in Redis with TTL; `webauthn` library verifies signature |
| **I**nformation disclosure | Private key exfiltration | Keys never leave authenticator; server stores public credential only |

### Admin panel

| Threat | Vector | Mitigation |
|---|---|---|
| **E**levation of privilege | Non-admin hits admin routes | `RequireSuperAdmin` middleware; audit log on sensitive mutations |

### Claude Code hooks

| Threat | Vector | Mitigation |
|---|---|---|
| **I**nformation disclosure | Hook reads stdin (event JSON) and forwards content | `eop-hook` (`backend/cmd/eop-hook`) parses only counts (chars/lines/lang), does not forward file content |
| **D**oS | Hook slows Claude Code | network error → stderr, exit 0; hook does not block the tool loop |

## Open issues

- [ ] Dedicated per-endpoint ingest rate limiter (Redis), on top of global Fiber 120/min.
- [ ] Audit log for `DELETE /v1/me/data` (who deleted, when) — V1.
- [ ] CSP on dashboard and `Content-Security-Policy` headers — V1.
- [ ] VS Code: migrate ingest token from `settings.json` to `SecretStorage` — V1.

## Re-validation

This document is living. Re-read before each release and when adding new components (e.g. a mobile app in V2 will need a separate STRIDE pass).
