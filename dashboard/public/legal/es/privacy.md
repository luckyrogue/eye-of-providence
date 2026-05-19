# Aviso de privacidad — Eye of Providence

**Última actualización:** 2026-05-19. Proveedor del servicio: maintainer individual
`main@rysdavletov.org`. Contacto para privacidad, GDPR/CCPA y seguridad — el mismo correo.

Este documento describe qué recopilamos, la base legal, cómo almacenamos los datos y cómo
puede ejercer sus derechos (incluidos equivalentes GDPR y CCPA). En instalaciones
self-hosted, los datos no salen de su infraestructura; este documento se aplica a la
instancia gestionada en `https://eop.rysdavletov.org`.

## 1. Qué recopilamos

### 1.1 Nunca sale del equipo del usuario

- **Contenido de archivos**, código fuente, archivos abiertos en el IDE.
- **Prompts en chats de IA** y **respuestas de modelos de IA** (la extensión del navegador
  sabe que está en una página de ChatGPT pero no lee el texto del diálogo).
- **Pulsaciones en bruto** — solo contadores, no secuencias.
- **Contenido del portapapeles** — solo sha256 + tamaño (véase §1.2).
- Títulos y contenido de ventanas en modos **privado/incógnito** y de listas
  negras definidas por el usuario.
- Contenido de **capturas de pantalla** — no las realizamos.

### 1.2 Qué se envía al backend

| Categoría | Concretamente | Para qué |
|---|---|---|
| Identificación | `user_id` (UUID), `device_id` (UUID), `session_id` | Vincular eventos a la cuenta |
| App en primer plano | bundle id (macOS) o nombre de proceso (Windows) | «En qué trabaja» |
| Duraciones | `duration_ms` de foco en una app | Tiempo activo / AFK |
| Entrada (contadores) | `chars_in` (conteo de pulsaciones, SIN contenido), `mouse_clicks` | Diferenciar manual vs asistido por IA |
| Huella del portapapeles | hash `sha256` + tamaño en bytes | Atribución de eventos paste (IA vs otro) |
| Canal de IA (si aplica) | provider (`openai`/`anthropic`/...), channel (`chat`/`inline`/`agent`/`cli`) | Desglose de uso de IA |
| Atribución de proyecto (opc.) | `project_id`, `file_lang` (solo tipo de archivo, no ruta) | Informes por proyecto |
| Metadatos de auth | email (login), login de GitHub (si OAuth), hashed_token para API keys | Auth |
| Informes | informe markdown generado (creado por modelo de IA a partir de agregados anteriores) | Historial |

### 1.3 Qué se envía a terceros

- **Google Gemini API** (`gemini-2.5-flash`) — recibe agregados numéricos
  para generar informes semanales/mensuales en texto. El prompt consiste en sus
  métricas agregadas (horas, %, apps principales). NO enviamos:
  ruta completa del bundle de la app, eventos en bruto ni nada de §1.1. Según
  [términos de Google AI](https://ai.google.dev/gemini-api/terms), el nivel gratuito
  puede usar prompts para mejorar modelos; el de pago — no. Véase §6.
- **Resend** (`api.resend.com`) — correo transaccional (verificación,
  restablecimiento de contraseña). Solo recibe dirección de correo + cuerpo del mensaje.
  [Privacidad de Resend](https://resend.com/legal/privacy).
- **GitHub** — si usa inicio de sesión OAuth, GitHub devuelve email
  + login. Véase [alcance OAuth de GitHub](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/scopes-for-oauth-apps).
- **GHCR (registro Docker)** — el agente de escritorio descarga actualizaciones vía
  el actualizador Tauri; ese tráfico va a GitHub.

## 2. Base legal (GDPR art. 6)

- **Art. 6(1)(b) Ejecución de un contrato** — tratamiento principal de sus datos
  para prestar el servicio (seguimiento de su propio tiempo).
- **Art. 6(1)(a) Consentimiento** — correos de marketing (aún no los enviamos) y
  envío de agregados a Google Gemini (opt-in explícito en
  Settings → AI Reports).
- **Art. 6(1)(f) Interés legítimo** — registros de seguridad (audit_log), supervisión
  de infracciones de rate-limit. Interés equilibrado — proteger el servicio.

Menores: el servicio no está destinado a personas menores de 16 años (o 13 en jurisdicciones
aplicables). No verificamos la edad; si detectamos un registro de menor,
eliminamos la cuenta de inmediato mediante `DELETE /v1/me/data`.

## 3. Retención

| Qué | Dónde | Plazo |
|---|---|---|
| Eventos (en bruto) | tabla ClickHouse `events` | TTL 18 meses, luego auto-drop por partición |
| Eventos de atribución (derivados) | ClickHouse `attribution_events` | TTL 18 meses |
| Perfil de usuario | Postgres `users` | Hasta eliminación de la cuenta |
| Audit log | Postgres `audit_log` | 24 meses |
| Informes (generados por IA) | Postgres `reports` | Hasta eliminación de la cuenta |
| Búfer SQLite local del agente | Disco local del usuario | 7 días (GC cada hora) |
| Logs del backend | stdout → agregador del hosting | 30 días |

## 4. Sus derechos

### 4.1 Acceso + portabilidad (art. 15, 20)

`GET /v1/me/export` (requiere Bearer JWT) devuelve JSON legible por máquina
con todos sus datos: perfil, devices, projects, consent, reports, API
tokens (sin `hashed_token`), historial completo de eventos (límite ~200k más recientes).

Panel: **Settings → Privacy → Export my data**.

### 4.2 Supresión / derecho al olvido (art. 17)

`DELETE /v1/me/data` borra:
- todos los eventos en ClickHouse (`ALTER ... DELETE WHERE user_id = ?`);
- reports + api_tokens + consent + projects + devices + fila de usuario en Postgres.

Panel: **Settings → Danger zone → Delete all my data**. La acción
es irreversible; el búfer SQLite local no se elimina (está en el equipo del usuario —
borrar manualmente con **Quit & wipe local data**).

### 4.3 Rectificación, limitación, oposición

Correo a `main@rysdavletov.org` con asunto `[GDPR DSAR]`. Plazo de respuesta —
30 días (art. 12(3)).

## 5. Seguridad

Véase [SECURITY.md](../.github/SECURITY.md). En resumen:
- Backend: bcrypt cost 10, JWT HS256 con revocación `token_version`, TTL 1h
  en restablecimiento de contraseña, rate-limit (10/min endpoints auth, 120/min /v1).
- Imagen: firmada (Cosign keyless), atestación SLSA L3, SBOM CycloneDX —
  verificada en self-host.
- Agente: búfer SQLite cifrado AES-256-GCM, clave en OS Keychain/DPAPI.

Respuesta a incidentes: acuse en 48 h, plazo de remediación 5 días laborables.

## 6. Subencargados

Lista completa de terceros que tratan datos:

| Subencargado | Finalidad | Ubicación | Acuerdo |
|---|---|---|---|
| Dokploy (si managed) | Hosting backend + BBDD | UE/EE. UU. (según deploy) | DPA bajo solicitud |
| Google (Gemini API) | Generación de informes con IA | EE. UU. | [DPA Google Cloud](https://cloud.google.com/terms/data-processing-addendum) |
| Resend | Correos transaccionales | EE. UU. | [DPA Resend](https://resend.com/legal/dpa) |
| GitHub | OAuth + GHCR | EE. UU. | [DPA GitHub](https://docs.github.com/en/site-policy/privacy-policies/global-privacy-practices) |

Los cambios en la lista se anuncian con 30 días de antelación en las release notes.

## 7. Transferencias internacionales

El hosting del backend depende de su elección self-host. Instancia gestionada —
Fráncfort, Alemania (UE). Transferencia a Google Gemini (EE. UU.) — bajo
[Cláusulas Contractuales Tipo](https://commission.europa.eu/law/law-topic/data-protection/international-dimension-data-protection/standard-contractual-clauses-scc_en).

## 8. Instancias self-hosted

En self-host (`docker-compose.full.yml`), los datos no salen de su
infraestructura salvo:
- API Gemini, si se define `EOP_GEMINI_API_KEY` (puede dejarse vacío).
- Resend, si se define `EOP_RESEND_API_KEY` (puede dejarse vacío).
- OAuth de GitHub, si se define `EOP_GITHUB_CLIENT_ID`.

El maintainer self-hosted es responsable del tratamiento independiente. Este documento
no le obliga; úselo como referencia.

## 9. Cambios

Este documento está versionado en git (`docs/privacy.md`). Los cambios
sustanciales se anuncian:
- en release notes (`CHANGELOG.md`);
- mediante notificación in-app en el panel (instancia gestionada);
- por correo a usuarios que aceptaron actualizaciones del producto.

## 10. Reclamaciones y autoridades

UE: puede presentar una reclamación ante la autoridad de control de su país
([lista completa](https://edpb.europa.eu/about-edpb/about-edpb/members_en)).
El maintainer no está en la UE; no hay representante art. 27 designado
(el servicio no se dirige sistemáticamente a residentes de la UE fuera del
segmento indie; lo revisaremos si se superan umbrales).

California (CCPA): solicitudes de divulgación / supresión / opt-out — mismo
correo.
