# Threat model — STRIDE

Скоуп: backend (`cmd/api`), desktop agent (Tauri), browser extension (MV3), VS Code plugin, Claude Code hooks.

## STRIDE по компонентам

### Backend (Go)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Поддельные events с чужим user_id | JWT перепривязывает event.user_id из токена в `ingest/handler.go:39`; client-supplied user_id игнорируется |
| **T**ampering | Подмена событий в транспорте | HTTPS на проде (см. `docs/self-hosting.md` production checklist) |
| **R**epudiation | "Я этого не отправлял" | `eop_ingest_events_*` метрики + access logs (Fiber middleware/logger) |
| **I**nformation disclosure | Утечка чужих данных через analytics | Все analytics-endpoints фильтруют по `claims.UserID`; cross-user query невозможна |
| **D**oS | Заливка событиями | Phase 8: rate-limit в Redis. Сейчас mitig: validEvent отбрасывает >24h durations |
| **E**levation of privilege | dev-token в production | `EOP_ENV=production` — disable dev-token endpoint (V1 TODO) |

### Desktop agent (Tauri)

| Threat | Vector | Mitigation |
|---|---|---|
| **S**poofing | Чужой процесс шлёт события через local API | Bearer token в `~/<data>/eop.local-token`, проверяется в `core/local_api.rs:handle` |
| **T**ampering | Модификация SQLite буфера | Локальный файл с правами user-only; шифрование at-rest — V1 TODO |
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

### Claude Code hooks

| Threat | Vector | Mitigation |
|---|---|---|
| **I**nformation disclosure | Hook читает stdin (event JSON) и пересылает | `eop-claude-hook.sh` НИКОГДА не читает stdin; только factual: category=ai, source=cli |
| **D**oS | Hook замедляет Claude Code | curl с `\|\| true` — hook никогда не блокирует; всегда exit 0 |

## Open issues

- [ ] Rate limiting на `/v1/ingest` (Redis) — V1.
- [ ] dev-token endpoint должен 404 в `EOP_ENV=production` — V1.
- [ ] Зашифровать SQLite буфер (`age`/AES-GCM ключ из Keychain/DPAPI) — V1.
- [ ] Audit log для `DELETE /v1/me/data` (кто удалил, когда) — V1.
- [ ] CSP на dashboard и `Content-Security-Policy` headers — V1.

## Re-validation

Этот документ — живой. Перечитывать перед каждым релизом и при добавлении новых компонентов (например, mobile app в V2 потребует отдельного STRIDE прохода).
