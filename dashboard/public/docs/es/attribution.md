# Atribución de código v2

Objetivo: precisión del 90 %+ en la clasificación «IA vs humano» por hunks/trazos. El
algoritmo v1 actual (tamaño de pegado + dominios de IA) ronda el 70 %. v2 añade 3
fuentes de señal:

1. **Aceptación inline VS Code/Cursor** — detección heurística por ráfagas
2. **Ediciones del agente Claude Code** — vía [Hooks API](https://docs.claude.com/en/docs/claude-code/hooks)
3. **Dominios de IA en el navegador** — claude.ai, chatgpt.com, perplexity, etc. (v1, sin cambios)

## Categorías y atribución útil

Dos capas de categorías:

- **Raw `events.category`** (ingest): `idle`, `manual`, `ai`, `reading`, `refactor`, `other`.
- **Derived `attribution_events.category`** (worker): `typed`, `pasted_ai`, `pasted_other`,
  `ai_inline`, `ai_agent`, `refactor`, `unknown` — véase
  [`002_attribution_events.up.sql`](../backend/internal/migrate/sql/clickhouse/002_attribution_events.up.sql).

La tabla siguiente mapea eventos **en bruto** a señales (antes de la fase B del worker):

| raw `category` | `ai_provider` | `ai_channel` | Señal |
|---|---|---|---|
| `manual` | — | — | escritura / pegado sin etiqueta de IA |
| `ai` | `copilot` | `inline` | ráfaga VS Code o pegado único >= umbral |
| `ai` | `cursor` | `inline` | Cursor (`vscode.env.appName === "Cursor"`) + misma heurística |
| `ai` | `claude-code` | `agent` | `eop-hook` en PostToolUse `Edit\|Write\|MultiEdit` |
| `ai` | `openai` / `anthropic` | `chat` | extensión: chatgpt.com / claude.ai |
| `refactor` | — | — | cambio estructural en IDE |
| `reading` | — | — | foco sin ediciones |

## Extensión VS Code / Cursor

Archivo: `ide-vscode/src/extension.ts`. Heurística doble:

1. **Inserción grande única** (`inserted >= 80 chars` en un `contentChange`) →
   `ai_inline`. «Aceptar sugerencia Copilot» clásico o pegado desde chat.
2. **Ráfaga** (varios `contentChange` en <100 ms, total >= 80) →
   también `ai_inline`. La finalización inline de Copilot/Cursor escribe
   1–3 caracteres por token; escritura normal — intervalo >150 ms entre pulsaciones.

`ai_provider` vía `vscode.env.appName`:
- `"Cursor"` → `cursor`
- si no → `copilot` (por defecto para sugerencias inline de VS Code)

`app_bundle` ligado al IDE: `com.todesktop.230313mzl4w4u92` (Cursor) o
`com.microsoft.VSCode`.

### Limitaciones conocidas

- El evento real Copilot Accept requiere `proposed API`
  (`commands.onWillExecuteCommand`) — aún no en VS Code estable.
  La heurística de ráfaga cubre 85 %+; el resto va como pegado.
- `aichat.acceptDiff` de Cursor no se intercepta por API pública — misma heurística.
- Si Copilot transmite muy lento (>100 ms entre caracteres), la detección de ráfaga
  puede fallar y atribuir como escritura. Ajuste `BURST_WINDOW_MS` si cambia el patrón.

## Hook de Claude Code

Señal más fuerte: Claude Code dispara el hook en cada
`Edit`/`Write`/`MultiEdit` — ambigüedad cero.

### Instalación

1. Compilar el binario del hook:

   ```sh
   cd backend && go build -o ~/.local/bin/eop-hook ./cmd/eop-hook
   ```

   (o `go install ./cmd/eop-hook` en `$GOBIN`).

2. Obtener token EOP (solo **desarrollo local**; en producción `dev-token` deshabilitado):

   ```sh
   # backend con EOP_ENABLE_DEV_TOKEN=true (por defecto en development)
   curl -X POST http://localhost:8080/v1/auth/dev-token | jq -r .token
   ```

   Guardar en `~/.zshrc` / `~/.bashrc`:

   ```sh
   export EOP_TOKEN="<token>"
   export EOP_BACKEND="http://localhost:8080"   # o su URL self-host
   ```

   En producción use token API del panel (`eop_<…>`) o emparejamiento de dispositivo.

3. Conectar el hook en ajustes de Claude Code (`~/.claude/settings.json` global
   o `.claude/settings.json` en el repo solo proyecto):

   ```json
   {
     "hooks": {
       "PostToolUse": [{
         "matcher": "Edit|Write|MultiEdit",
         "hooks": [{ "type": "command", "command": "eop-hook" }]
       }]
     }
   }
   ```

4. Reinicie la sesión de Claude Code. Cada Edit/Write envía evento con
   `category=ai`, `ai_provider=claude-code`, `ai_channel=agent`.

### Qué se cuenta

- `Write` — longitud total de `content` en `chars_in`, conteo `\n` → `lines_added`.
- `Edit` — longitud de `new_string` en caracteres, `\n` en `new_string` → `lines_added`,
  `\n` en `old_string` → `lines_removed`.
- `MultiEdit` — suma de todas las ediciones.
- `Read` / `Bash` / `Grep` — ignorados (sin significado de atribución).

### Privacidad

El hook envía solo conteos (chars/lines/lang), **no contenido** de archivos. Validación en ingest —
`domain.ValidEvent()` en
[`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go);
sin campos de texto en el formato wire.

### Modo de fallo

Si `EOP_TOKEN` no está definido — el hook sale en silencio (exit 0). Error de red —
stderr, pero no bloquea el flujo de Claude Code. Por diseño: no romper el ciclo de herramientas
si el backend no está disponible.

## Extensión del navegador

Archivo: `browser-extension/src/entrypoints/background.ts`. Sin cambios en v2 — rastrea
foco en dominios de IA, emite `category=ai`, `ai_provider` del mapping
`ai-domains.ts`, `ai_channel=chat` (no inline).

## Agente Tauri

Archivo: `agent/src-tauri/src/core/`. Emite eventos de foco en aplicaciones nativas
(VS Code, Cursor, Claude Desktop). Para atribución de IA depende de la extensión VS Code /
del navegador en paralelo. v2 sin cambios.

## Hoja de ruta

- **Apply** para Cursor: comando envoltorio `eop.captureCursorApply`,
  override `cmd+enter` → apply de Cursor + evento `ai` explícito.
- **Copilot Accept**: esperar `commands.onWillExecuteCommand` estable o
  API `vscode.copilot` propuesta. En estable actual la heurística de ráfaga es lo mejor.
- **Plugin JetBrains**: análogo de VS Code para IntelliJ
  (PyCharm/WebStorm/etc.). Sprint grande, aplazado.
