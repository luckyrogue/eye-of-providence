# Қауіп моделі — STRIDE

> **Күйі:** ⚠ соңғы тексеру 2026-05-04 — alpha → beta
> көтеру алдында қайта қарау қажет. Белгілі бос орындар (соңғы тексеруден кейін қосылды): passkey/WebAuthn
> аутентификация, GDPR-export endpoint (`GET /v1/me/export`), admin panel (super-admin
> агрегаттар + audit log көрінісі), үшінші тарап интеграциялары (Resend email,
> Dokploy hosting). [`tech-debt.md`](tech-debt.md) C8 ішінде бақыланады.

Қамту: backend (`cmd/api`), desktop agent (Tauri), browser extension (MV3), VS Code plugin, Claude Code hooks.

## Компоненттер бойынша STRIDE

### Backend (Go)

| Қауіп | Вектор | Шешу |
|---|---|---|
| **S**poofing | Басқа user_id бар жалған events | JWT `ingestapp/service.go` ішінде event.user_id мәнін токеннен қайта байлайды (`e.UserID = userID`); client жіберген user_id елемейді |
| **T**ampering | Транзитте events өзгерту | Production-да HTTPS (қараңыз `docs/self-hosting.md` production checklist) |
| **R**epudiation | «Мен жібермедім» | `eop_ingest_events_*` метрикалары + access logs (Fiber middleware/logger) |
| **I**nformation disclosure | Analytics арқылы басқа пайдаланушы деректерінің ашылуы | Барлық analytics endpoint-тері `claims.UserID` бойынша сүзіледі; cross-user query мүмкін емес |
| **D**oS | Events толтыру | Fiber limiter `/v1/*` үшін 120 req/min (`cmd/api/main.go`); `domain.ValidEvent` >24h duration-ды тастайды; dedicated per-ingest Redis limiter — roadmap |
| **E**levation of privilege | production-да dev-token | `EOP_ENV=production` кезінде `EOP_ENABLE_DEV_TOKEN` тыйым салынған; өшірілгенде route 404 (`config.go`, `dev_token_test.go`) |

### Desktop agent (Tauri)

| Қауіп | Вектор | Шешу |
|---|---|---|
| **S**poofing | Бөтен процесс local API арқылы events жібереді | Bearer token `~/<data>/eop.local-token` ішінде, `core/local_api.rs:handle` тексереді |
| **T**ampering | SQLite buffer өзгерту | User-only рұқсаттары бар локальді файл; AES-256-GCM at-rest (`agent/src-tauri/src/core/crypto.rs`, `store.rs`) |
| **R**epudiation | — | Тек local, multi-user жоқ |
| **I**nformation disclosure | Файл мазмұны / prompt жинау | **Архитектуралық инвариант**: agent Claude Code hook stdin оқымайды, файл body парс етпейді; тек timestamps, char counts, hashes. Бұзу = bug |
| **D**oS | event_buffer арқылы disk толтыру | TTL және batch flush; Phase 8-де — pending_count hard limit |
| **E**levation of privilege | macOS Accessibility рұқсаты | Onboarding flow арқылы анық сұралады; онсыз keystroke counts жоқ (graceful degradation) |

### Browser extension (MV3)

| Қауіп | Вектор | Шешу |
|---|---|---|
| **S**poofing | DOM-да content-script алдау | content script тек selection size + host оқиды; мазмұн сериализацияланбайды |
| **T**ampering | Бетте XSS → жалған events | Events `chrome.runtime.sendMessage` арқылы — sender service worker тексереді (host whitelist) |
| **I**nformation disclosure | URL/title/content кездейсоқ жіберу | `host_permissions` whitelist (тек AI домендері + localhost); content scripts тек `host` + `size` жібереді |
| **E**levation of privilege | Extension арқылы OAuth cookie ұрлау | Extension бөтен cookie store-ға қол жеткізе алмайды; JWT `chrome.storage.local` ішінде (extension бойынша оқшауланған) |

### VS Code extension

| Қауіп | Вектор | Шешу |
|---|---|---|
| **I**nformation disclosure | Diff арқылы файл мазмұны | `onDidChangeTextDocument` ұзындық пен timestamp береді; нақты мәтін payload-қа кірмейді (`extension.ts::onChange`) |
| **T**ampering | settings.json-дағы token — file access бар кез келген адам | Кейінірек ауыстыру: `secrets.SecretStorage` API (V1) |

### WebAuthn / passkeys

| Қауіп | Вектор | Шешу |
|---|---|---|
| **S**poofing | Credential replay | Challenge Redis-те TTL-пен сақталады; `webauthn` library қолтаңбаны тексереді |
| **I**nformation disclosure | Private key exfil | Кілттер authenticator-дан шықпайды; server тек public credential сақтайды |

### Admin panel

| Қауіп | Вектор | Шешу |
|---|---|---|
| **E**levation of privilege | Admin емес admin route-қа кіру | `RequireSuperAdmin` middleware; сезімтал mutation-дарда audit log |

### Claude Code hooks

| Қауіп | Вектор | Шешу |
|---|---|---|
| **I**nformation disclosure | Hook stdin (event JSON) оқып мазмұн жібереді | `eop-hook` (`backend/cmd/eop-hook`) тек counts (chars/lines/lang) парс етеді, файл мазмұнын жібермейді |
| **D**oS | Hook Claude Code-ты баяулатады | network error → stderr, exit 0; hook tool циклін блоктамайды |

## Ашық мәселелер

- [ ] Dedicated per-endpoint ingest rate limiter (Redis), жалпы Fiber 120/min үстіне.
- [ ] `DELETE /v1/me/data` үшін audit log (кім өшірді, қашан) — V1.
- [ ] Dashboard CSP және `Content-Security-Policy` headers — V1.
- [ ] VS Code: ingest token `settings.json` → `SecretStorage` көшіру — V1.

## Қайта тексеру

Бұл құжат тірі. Әр релиз алдында және жаңа компонент қосқанда қайта оқыңыз (мысалы, V2 mobile app бөлек STRIDE өткізуді қажет етеді).
