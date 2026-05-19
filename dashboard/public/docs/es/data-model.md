# Modelo de datos

Dos almacenes: **Postgres** para estado transaccional (users, teams, devices,
projects, reports, audit log) y **ClickHouse** para analítica (events,
attribution, agregados materializados). Fuente de verdad del formato wire de ingest —
[`docs/api/openapi.yaml`](api/openapi.yaml) (`components.schemas.Event`) y
tipo Go [`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go).
Transporte: JSON `POST /v1/ingest` (no protobuf/gRPC).

Todas las migraciones SQL están versionadas en `backend/internal/migrate/sql/`
(`postgres/NNN_*.up.sql` y `clickhouse/NNN_*.up.sql`). La API las aplica
automáticamente al arrancar si `EOP_AUTO_MIGRATE=true`.

## 1. Postgres

### 1.1 Identidad y propiedad

| Tabla | Finalidad | Notable |
|---|---|---|
| `users` | Cuentas + soft-delete (`deleted_at`) | `email UNIQUE`, OAuth `github_login`, FK `team_id`, `token_version` para revocación |
| `teams` | Grupos + plan (`free`/`pro`/...) | `settings_json JSONB` config flexible por equipo |
| `team_members` | Muchos a muchos users↔teams + rol | Véase `003_multi_team_projects.up.sql` |
| `webauthn_credentials` | Credenciales FIDO2/passkey | Véase `020_webauthn_credentials.up.sql` |

### 1.2 Devices y projects

| Tabla | Campo | Por qué |
|---|---|---|
| `devices` | `(user_id, fingerprint) UNIQUE` | Vinculación del agente de escritorio; el pair-flow crea fila aquí |
| `projects` | `root_path_hash` (sha256), no la ruta | Atribución por proyecto sin revelar FS |

### 1.3 Consent y API tokens

- `consent` — marcas granted/revoked por ámbito (`telemetry`, `ai_reports`, …).
- `api_tokens` — bearer para CI/scripts; `hashed_token TEXT NOT NULL UNIQUE`
  (bcrypt). El texto plano se devuelve una sola vez al crear.

### 1.4 Reports y audit

- `reports` — informes markdown generados por IA, `period` ∈ `weekly_2026_W18`/`monthly_2026_05`,
  `prompt_version` para seguimiento A/B de prompts.
- `audit_log` (migración `016_audit_log.up.sql`) — eventos sensibles:
  `login_failed`, `password_reset`, `token_created/revoked`,
  `user_deleted`. Retención 24 meses.

### 1.5 Webhooks y rate-limit

- `webhooks` (`011_webhooks.up.sql`) — notificaciones HTTP salientes (Slack/
  Discord) con reintentos y secreto HMAC.
- Estado de rate-limit — en Redis, no Postgres. Conteos fallidos van a
  `audit_log` para post-mortem.

### 1.6 Cascade y retención

`ON DELETE CASCADE` en todas las FK a `users(id)` — `DELETE /v1/me/data`
elimina perfil + devices + projects + consent + reports +
api_tokens con un `DELETE FROM users WHERE id = $1`. Soft-delete
(`users.deleted_at`) solo bloquea login, no umbral de retención.

## 2. ClickHouse

### 2.1 Eventos en bruto — `events`

```
events (ts, user_id, device_id, session_id,
        app_bundle, category, source, ai_provider, ai_channel,
        project_id, file_lang,
        duration_ms, chars_in, lines_added, lines_removed,
        meta)
ORDER BY (user_id, ts)
PARTITION BY toYYYYMM(ts)
TTL toDateTime(ts) + INTERVAL 18 MONTH
```

- `category` ∈ `idle`, `manual`, `ai`, `reading`, `refactor`, `other`.
- `source` ∈ `os`, `browser`, `ide`, `cli`.
- `meta String` — JSON con extensiones seguras: `clipboard_sha256`,
  `clipboard_bytes`, `mouse_clicks`. Sin contenido (véase [/privacy](/privacy) §1.1).
- TTL 18 meses — auto-drop por partición. No usamos `DELETE WHERE` para
  retención (costoso); solo para `DELETE /v1/me/data`.

### 2.2 Atribución — `attribution_events`

Derivado de `events` vía `attribution worker` (`backend/internal/attribution/`).
Categorías más precisas que en bruto:

```
category Enum8(
    'typed'        = 1,
    'pasted_ai'    = 2,
    'pasted_other' = 3,
    'ai_inline'    = 4,
    'ai_agent'     = 5,
    'refactor'     = 6,
    'unknown'      = 7
)
```

El worker corre por lotes. Fases (véase
[`backend/internal/attribution/worker.go`](../backend/internal/attribution/worker.go)):

**Fase A (ahora):**
1. lee `events` recientes con `category in (manual, ai)`;
2. mapea cada evento a `attribution_events.category` por reglas planas;
3. escribe lote en `attribution_events`, idempotente por dedup `(ts, user_id, project_id)`.

**Fase B (hoja de ruta):**
1. une paste `clipboard_sha256` con `clipboard_signal` — ventana 30 s;
2. atribución por hunk vía snapshots diff en extensión IDE;
3. distinción inline/agente precisa donde hay hooks API directos.

### 2.3 Agregados materializados

MVs en cascada (`backend/internal/migrate/sql/clickhouse/003_*`, `004_*`):

```
events (raw)  →  events_hourly_agg  →  events_daily_agg
              events_hourly_mv       events_daily_mv
```

- `events_hourly_agg` — `SummingMergeTree` en `(user_id, bucket_ts, category, file_lang)`.
- `events_daily_agg` — bucket diario para rangos largos (30/90 días).

**Enrutado** (`store/clickhouse.go`): `AggregateByCategory` / bulk → MV **daily** si
`since ≥ 30d`, si no hourly.

**Reducción** (100 usuarios activos): raw ~10M filas/día → hourly ~290K (**35×**) → daily ~12K
(**840×** vs raw).

**Benchmark** (`go run ./cmd/ch-bench` desde `backend/`):

```sh
EOP_CH_DSN="clickhouse://default:@localhost:19000/eop_test" \
  go run ./cmd/ch-bench --users=10 --days=30 --events-per-user-per-day=10000
```

### 2.4 Caché de lectura Redis (`store/cached.go`)

Decorador sobre `EventStore` si Redis está disponible. TTLs: `AggregateByCategory` 10m,
etc. Redis caído → miss + log, sin fallo.

## 3. Flujo de eventos

```
┌──────────┐  events     ┌──────────┐   raw insert    ┌────────────┐
│  Agent   │ ──────────► │  /v1/    │ ──────────────► │ events     │
│ (desktop,│  batched    │ ingest   │   ClickHouse    │ MergeTree  │
│ browser, │  HTTP+JSON  │ handler  │                 └─────┬──────┘
│ IDE, CLI)│             │ (Go)     │                       │ MV trigger
└──────────┘             └──────────┘                       ▼
                                                    ┌────────────────────┐
                                                    │ events_hourly_agg  │
                                                    │ events_daily_agg   │
                                                    └────────────────────┘

                                ┌─────────────────┐  batch SELECT
                                │ attribution-    │ ────────────► events
                                │ worker (Go)     │
                                └────────┬────────┘
                                         ▼
                                 ┌────────────────────┐
                                 │ attribution_events │
                                 └────────────────────┘
```

## 4. Invariantes transversales

- **`user_id` siempre UUIDv4**, nunca email ni enviado por cliente. El backend
  sobrescribe `event.user_id` desde claims JWT (`ingestapp/service.go`).
- **Sin contenido**. Tablas con contadores, hashes o enums. `meta` —
  solo claves seguras. Más en [/privacy](/privacy) §1.1.
- **Migraciones idempotentes**. `001..N` con advisory_lock + `IF NOT EXISTS`.
- **TTL — único mecanismo de retención** en analítica. Borrado GDPR — explícito
  `ALTER TABLE ... DELETE WHERE user_id = ?`.

## 5. Extensión

Al añadir campo a Event:
1. Actualizar `components.schemas.Event` en [`docs/api/openapi.yaml`](api/openapi.yaml).
2. Añadir en `backend/internal/ingest/domain/event.go` y `backend/internal/store/event.go`.
3. Columna vía migración `005_*.up.sql`.
4. Actualizar `Insert()` en `clickhouse.go`.
5. Si agrega — nueva MV en `events_hourly_agg`.
6. Agente: `agent/src-tauri/src/core/event.rs`.

Para extensiones sin contenido preferir `meta map<string,string>`
(ya existe; sin migración).
