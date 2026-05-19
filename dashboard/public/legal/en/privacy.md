# Privacy Notice — Eye of Providence

**Last updated:** 2026-05-19. Service provider: individual maintainer
`main@rysdavletov.org`. Contact for privacy, GDPR/CCPA, and security — same email.

This document describes what we collect, the legal basis, how we store data, and how
you can exercise your rights (including GDPR and CCPA equivalents). On self-hosted
installations, data does not leave your infrastructure; this document applies to the
managed instance at `https://eop.rysdavletov.org`.

## 1. What we collect

### 1.1 Never leaves the user's machine

- **File contents**, source code, files open in the IDE.
- **AI chat prompts** and **AI model responses** (the browser extension knows
  you are on a ChatGPT page but does not read dialogue text).
- **Raw keystrokes** — we store counters only, not sequences.
- **Clipboard contents** — only sha256 + size (see §1.2).
- Window titles and content in **private/incognito** modes and from user-defined
  blacklist.
- **Screenshot** content — we do not capture screenshots.

### 1.2 What is sent to the backend

| Category | Specifically | Purpose |
|---|---|---|
| Identification | `user_id` (UUID), `device_id` (UUID), `session_id` | Tie events to an account |
| Foreground app | bundle id (macOS) or process name (Windows) | "What you are working on" |
| Durations | `duration_ms` of focus in an app | Active time / AFK |
| Input (counters) | `chars_in` (keystroke count, NO content), `mouse_clicks` | Manual vs AI-assisted differentiation |
| Clipboard fingerprint | `sha256` hash + size in bytes | Attribution of paste events (AI vs other) |
| AI channel (if applicable) | provider (`openai`/`anthropic`/...), channel (`chat`/`inline`/`agent`/`cli`) | AI usage breakdown |
| Project attribution (opt.) | `project_id`, `file_lang` (file type only, not path) | Per-project reports |
| Auth metadata | email (for login), GitHub login (if OAuth), hashed_token for API keys | Auth |
| Reports | generated markdown report (built by AI model from aggregates above) | History |

### 1.3 What is sent to third parties

- **Google Gemini API** (`gemini-2.5-flash`) — receives numeric aggregates
  to generate weekly/monthly text reports. The prompt consists of your aggregated
  metrics (hours, %, top apps). We do NOT send: full app bundle paths, raw events,
  or anything from §1.1. Per
  [Google AI terms](https://ai.google.dev/gemini-api/terms), the free tier
  may use prompts to improve models; paid tier — not. See §6.
- **Resend** (`api.resend.com`) — transactional email (verification,
  password reset). Receives only email address + message body.
  [Resend privacy](https://resend.com/legal/privacy).
- **GitHub** — if you use OAuth login, GitHub returns email
  + login. See [GitHub OAuth scope](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps).
- **GHCR (Docker registry)** — the desktop agent downloads updates via
  the Tauri updater; that traffic goes to GitHub.

## 2. Legal basis (GDPR Art. 6)

- **Art. 6(1)(b) Performance of a contract** — core processing of your data
  to provide the service (tracking your own time).
- **Art. 6(1)(a) Consent** — marketing emails (we do not send them yet), and
  sending aggregates to Google Gemini (explicit opt-in in
  Settings → AI Reports).
- **Art. 6(1)(f) Legitimate interest** — security logs (audit_log), monitoring
  rate-limit violations. Balancing interest — protecting the service.

Children: the service is not intended for persons under 16 (or 13 in applicable
jurisdictions). We do not verify age; if we discover a minor's registration,
we delete the account immediately via `DELETE /v1/me/data`.

## 3. Retention

| What | Where | Period |
|---|---|---|
| Events (raw) | ClickHouse `events` table | 18 months TTL, then auto-drop by partition |
| Attribution events (derived) | ClickHouse `attribution_events` | 18 months TTL |
| User profile | Postgres `users` | Until account deletion |
| Audit log | Postgres `audit_log` | 24 months |
| Reports (AI-generated) | Postgres `reports` | Until account deletion |
| Local SQLite agent buffer | User's local disk | 7 days (GC every hour) |
| Backend logs | stdout → hosting aggregator | 30 days |

## 4. Your rights

### 4.1 Access + portability (Art. 15, 20)

`GET /v1/me/export` (requires Bearer JWT) returns machine-readable JSON
with all your data: profile, devices, projects, consent, reports, API
tokens (without `hashed_token`), full event history (cap ~200k most recent).

Dashboard: **Settings → Privacy → Export my data**.

### 4.2 Erasure / right to be forgotten (Art. 17)

`DELETE /v1/me/data` deletes:
- all events in ClickHouse (`ALTER ... DELETE WHERE user_id = ?`);
- reports + api_tokens + consent + projects + devices + user row in Postgres.

Dashboard: **Settings → Danger zone → Delete all my data**. This action
is irreversible; the local SQLite buffer is not deleted (it is on the user's machine —
clear manually via **Quit & wipe local data**).

### 4.3 Rectification, restriction, objection

Email `main@rysdavletov.org` with subject `[GDPR DSAR]`. Response time —
30 days (Art. 12(3)).

## 5. Security

See [SECURITY.md](../.github/SECURITY.md). In brief:
- Backend: bcrypt cost 10, JWT HS256 with `token_version` revocation, 1h TTL
  on password reset, rate-limit (10/min auth endpoints, 120/min /v1).
- Image: signed (Cosign keyless), SLSA L3 attestation, CycloneDX SBOM —
  verified on self-host.
- Agent: SQLite buffer encrypted AES-256-GCM, key in OS Keychain/DPAPI.

Incident response: 48h acknowledgement, 5 business days remediation timeline.

## 6. Sub-processors

Full list of third parties processing data:

| Sub-processor | Purpose | Location | Agreement |
|---|---|---|---|
| Dokploy (if managed) | Hosting backend + DBs | EU/US (depends on deploy) | DPA on request |
| Google (Gemini API) | AI report generation | US | [Google Cloud DPA](https://cloud.google.com/terms/data-processing-addendum) |
| Resend | Transactional emails | US | [Resend DPA](https://resend.com/legal/dpa) |
| GitHub | OAuth + GHCR | US | [GitHub DPA](https://docs.github.com/en/site-policy/privacy-policies/global-privacy-practices) |

List changes are announced 30 days in advance via release notes.

## 7. International transfers

Backend hosting depends on your self-host choice. Managed instance —
Frankfurt, Germany (EU). Transfer to Google Gemini (US) — under
[Standard Contractual Clauses](https://commission.europa.eu/law/law-topic/data-protection/international-dimension-data-protection/standard-contractual-clauses-scc_en).

## 8. Self-hosted instances

On self-host (`docker-compose.full.yml`), data does not leave your
infrastructure except for:
- Gemini API, if `EOP_GEMINI_API_KEY` is set (can be left empty).
- Resend, if `EOP_RESEND_API_KEY` is set (can be left empty).
- GitHub OAuth, if `EOP_GITHUB_CLIENT_ID` is set.

The self-hosted maintainer is an independent data controller. This document
does not bind you; use it as a baseline.

## 9. Changes

This document is versioned in git (`docs/privacy.md`). Material
changes are announced:
- via release notes (`CHANGELOG.md`);
- via in-app notification on the dashboard (managed instance);
- by email to users who agreed to product updates.

## 10. Complaints and regulators

EU: you may lodge a complaint with the supervisory authority in your country
([full list](https://edpb.europa.eu/about-edpb/about-edpb/members_en)).
The maintainer is not based in the EU; no Art. 27 representative is appointed
(the service is not systematically targeted at EU residents beyond the
indie segment; we will revisit if thresholds are exceeded).

US California (CCPA): disclosure / deletion / opt-out requests — same
email.
