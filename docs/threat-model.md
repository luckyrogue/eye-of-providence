# Threat model — STRIDE

Скоуп: backend (`cmd/api`), desktop agent (Tauri), browser extension (MV3), VS Code plugin, Claude Code hooks.

## STRIDE по компонентам

### Backend (Go)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Поддельные events с чужим user_id | JWT перепривязывает event.user_id из токена в `ingestapp/service.go` (`e.UserID = userID`); client-supplied user_id игнорируется |
| **T**ampering | Подмена событий в транспорте | HTTPS на проде (см. `docs/self-hosting.md` production checklist) |
| **R**epudiation | "Я этого не отправлял" | `eop_ingest_events_*` метрики + access logs (Fiber middleware/logger) |
| **I**nformation disclosure | Утечка чужих данных через analytics | Все analytics-endpoints фильтруют по `claims.UserID`; cross-user query невозможна |
| **D**oS | Заливка событиями | Fiber limiter 120 req/min на `/v1/*` (`cmd/api/main.go`); `domain.ValidEvent` отбрасывает >24h durations; dedicated per-ingest Redis limiter — roadmap |
| **E**levation of privilege | dev-token в production | `EOP_ENABLE_DEV_TOKEN` запрещён в `EOP_ENV=production`; route 404 когда disabled (`config.go`, `dev_token_test.go`) |

### Desktop agent (Tauri)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Чужой процесс шлёт события через local API | Bearer token в `~/<data>/eop.local-token`, проверяется в `core/local_api.rs:handle` |
| **T**ampering | Модификация SQLite буфера | Локальный файл с правами user-only; AES-256-GCM at-rest (`agent/src-tauri/src/core/crypto.rs`, `store.rs`) |
| **R**epudiation | — | Local-only, нет multi-user |
| **I**nformation disclosure | Сбор контента файлов / промптов | **Архитектурный invariant**: agent не читает stdin claude code хуков, не парсит body файлов; только timestamps, char counts, hashes. Нарушение — баг |
| **D**oS | Disk fill через event_buffer | TTL и batch flush; В Phase 8 — hard limit на pending_count |
| **E**levation of privilege | macOS Accessibility permission | Запрашивается явно через onboarding flow; без него keystroke counts недоступны (graceful degradation) |

### Browser extension (MV3)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Подмена content-script в DOM | content script читает только selection size + host; контент не сериализуется |
| **T**ampering | XSS на странице → инжект fake events | События проходят через `chrome.runtime.sendMessage` — sender проверяется service worker'ом (host whitelist) |
| **I**nformation disclosure | Случайная отправка URL/title/content | `host_permissions` whitelisted (только AI-домены + localhost); content scripts шлют ТОЛЬКО `host` + `size` |
| **E**levation of privilege | OAuth-cookie кража через extension | Extension не имеет доступа к чужим cookie store; JWT в `chrome.storage.local` (изолировано per-extension) |

### VS Code extension

| Threat | Vector | Mitigation |
|---|---|---|
| **I**nformation disclosure | Содержимое файла через diff | `onDidChangeTextDocument` даёт нам lengths и timestamps; реальный текст в наш payload не попадает (см. `extension.ts::onChange`) |
| **T**ampering | Конфиг token в settings.json — кто угодно с file access | Дальнейшее замещение: использовать `secrets.SecretStorage` API (V1) |

### WebAuthn / passkeys

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Credential replay | Challenge stored in Redis with TTL; `webauthn` library verifies signature |
| **I**nformation disclosure | Private key exfil | Keys never leave authenticator; server stores public credential only |

### Admin panel

| Threat | Vector | Mitigation |
|---|---|---|
| **E**levation of privilege | Non-admin hits admin routes | `RequireSuperAdmin` middleware; audit log on sensitive mutations |

### Claude Code hooks

| Threat | Vector | Mitigation |
|---|---|---|
| **I**nformation disclosure | Hook читает stdin (event JSON) и пересылает контент | `eop-hook` (`backend/cmd/eop-hook`) парсит только counts (chars/lines/lang), не пересылает file content |
| **D**oS | Hook замедляет Claude Code | network error → stderr, exit 0; hook не блокирует tool-цикл |

## Open issues

- [ ] Dedicated per-endpoint ingest rate limiter (Redis), поверх общего Fiber 120/min.
- [ ] Audit log для `DELETE /v1/me/data` (кто удалил, когда) — V1.
- [ ] CSP на dashboard и `Content-Security-Policy` headers — V1.
- [ ] VS Code: migrate ingest token from `settings.json` to `SecretStorage` — V1.

## Re-validation

Этот документ — живой. Перечитывать перед каждым релизом и при добавлении новых компонентов (например, mobile app в V2 потребует отдельного STRIDE прохода).
