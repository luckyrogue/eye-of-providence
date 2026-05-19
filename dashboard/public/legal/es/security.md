# Modelo de amenazas — STRIDE

> **Estado:** ⚠ última revisión 2026-05-04 — se requiere nueva revisión antes de pasar de alpha a beta.
> Brechas conocidas (añadidas tras la última revisión): autenticación passkey/WebAuthn,
> endpoint de exportación GDPR (`GET /v1/me/export`), panel de administración (vista super-admin de agregados + audit log),
> integraciones de terceros (correo Resend, hosting Dokploy). Registrado en [`tech-debt.md`](tech-debt.md) C8.

Alcance: backend (`cmd/api`), agente de escritorio (Tauri), extensión de navegador (MV3), plugin de VS Code, hooks de Claude Code.

## STRIDE por componente

### Backend (Go)

| Amenaza | Vector | Mitigación |
|---|---|---|
| **S**poofing | Eventos falsos con user_id ajeno | JWT reasigna event.user_id desde el token en `ingestapp/service.go` (`e.UserID = userID`); se ignora el user_id enviado por el cliente |
| **T**ampering | Manipulación de eventos en tránsito | HTTPS en producción (véase la checklist de producción en `docs/self-hosting.md`) |
| **R**epudiation | «Yo no envié eso» | Métricas `eop_ingest_events_*` + access logs (Fiber middleware/logger) |
| **I**nformation disclosure | Filtración de datos ajenos vía analytics | Todos los endpoints de analytics filtran por `claims.UserID`; consultas entre usuarios son imposibles |
| **D**oS | Inundación de eventos | Limiter Fiber 120 req/min en `/v1/*` (`cmd/api/main.go`); `domain.ValidEvent` descarta duraciones >24h; limiter Redis dedicado por ingest — roadmap |
| **E**levation of privilege | dev-token en producción | `EOP_ENABLE_DEV_TOKEN` prohibido con `EOP_ENV=production`; la ruta devuelve 404 cuando está deshabilitado (`config.go`, `dev_token_test.go`) |

### Agente de escritorio (Tauri)

| Amenaza | Vector | Mitigación |
|---|---|---|
| **S**poofing | Proceso ajeno envía eventos por API local | Bearer token en `~/<data>/eop.local-token`, verificado en `core/local_api.rs:handle` |
| **T**ampering | Modificación del búfer SQLite | Archivo local con permisos solo de usuario; AES-256-GCM en reposo (`agent/src-tauri/src/core/crypto.rs`, `store.rs`) |
| **R**epudiation | — | Solo local, sin multi-usuario |
| **I**nformation disclosure | Recogida de contenido de archivos / prompts | **Invariante arquitectónico**: el agente no lee stdin de hooks de Claude Code ni parsea cuerpos de archivos; solo timestamps, conteos de caracteres y hashes. Violación = bug |
| **D**oS | Llenado de disco vía event_buffer | TTL y flush por lotes; en la fase 8 — límite duro de pending_count |
| **E**levation of privilege | Permiso de Accesibilidad en macOS | Se solicita explícitamente en el onboarding; sin él no hay conteos de pulsaciones (degradación gradual) |

### Extensión de navegador (MV3)

| Amenaza | Vector | Mitigación |
|---|---|---|
| **S**poofing | Suplantación de content-script en el DOM | el content script solo lee tamaño de selección + host; no se serializa contenido |
| **T**ampering | XSS en la página → inyección de eventos falsos | Los eventos pasan por `chrome.runtime.sendMessage`; el service worker verifica el remitente (whitelist de hosts) |
| **I**nformation disclosure | Envío accidental de URL/título/contenido | `host_permissions` en whitelist (solo dominios de IA + localhost); los content scripts envían SOLO `host` + `size` |
| **E**levation of privilege | Robo de cookies OAuth vía extensión | La extensión no accede a almacenes de cookies ajenos; JWT en `chrome.storage.local` (aislado por extensión) |

### Extensión VS Code

| Amenaza | Vector | Mitigación |
|---|---|---|
| **I**nformation disclosure | Contenido de archivo vía diff | `onDidChangeTextDocument` aporta longitudes y timestamps; el texto real no entra en nuestro payload (véase `extension.ts::onChange`) |
| **T**ampering | Token en settings.json — cualquiera con acceso al archivo | Sustitución prevista: API `secrets.SecretStorage` (V1) |

### WebAuthn / passkeys

| Amenaza | Vector | Mitigación |
|---|---|---|
| **S**poofing | Replay de credencial | Challenge en Redis con TTL; la librería `webauthn` verifica la firma |
| **I**nformation disclosure | Exfiltración de clave privada | Las claves no salen del autenticador; el servidor solo guarda la credencial pública |

### Panel de administración

| Amenaza | Vector | Mitigación |
|---|---|---|
| **E**levation of privilege | No-admin accede a rutas admin | Middleware `RequireSuperAdmin`; audit log en mutaciones sensibles |

### Hooks de Claude Code

| Amenaza | Vector | Mitigación |
|---|---|---|
| **I**nformation disclosure | El hook lee stdin (JSON del evento) y reenvía contenido | `eop-hook` (`backend/cmd/eop-hook`) parsea solo conteos (chars/lines/lang), no reenvía contenido de archivos |
| **D**oS | El hook ralentiza Claude Code | error de red → stderr, exit 0; el hook no bloquea el ciclo de herramientas |

## Temas abiertos

- [ ] Rate limiter de ingest dedicado por endpoint (Redis), además del Fiber global 120/min.
- [ ] Audit log para `DELETE /v1/me/data` (quién eliminó, cuándo) — V1.
- [ ] CSP en el dashboard y cabeceras `Content-Security-Policy` — V1.
- [ ] VS Code: migrar token de ingest de `settings.json` a `SecretStorage` — V1.

## Revalidación

Este documento está vivo. Releerlo antes de cada release y al añadir componentes nuevos (p. ej. una app móvil en V2 requerirá un paso STRIDE aparte).
