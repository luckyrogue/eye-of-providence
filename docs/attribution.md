# Code attribution v2

Цель: точность 90%+ на классификации «AI vs human» по hunks/strokes. Текущий
algoritm v1 (paste size + AI domains) даёт ~70%. v2 добавляет 3 источника
сигналов:

1. **VS Code/Cursor inline accept** — heuristic burst-detection
2. **Claude Code agent edits** — через [Hooks API](https://docs.claude.com/en/docs/claude-code/hooks)
3. **Browser AI domains** — claude.ai, chatgpt.com, perplexity, etc. (v1, без изменений)

## Категории и полезная атрибуция

ClickHouse `attribution_events` (`backend/internal/migrate/sql/clickhouse/002_attribution_events.up.sql`):

| `category` | `ai_provider` | `ai_channel` | Сигнал |
|---|---|---|---|
| `manual` | — | — | typing < pasteThreshold (80 chars) |
| `ai` | `copilot` | `inline` | VS Code burst или single paste >= threshold |
| `ai` | `cursor` | `inline` | Cursor (vscode.env.appName === "Cursor") + same heuristic |
| `ai` | `claude-code` | `agent` | PostToolUse hook от Claude Code |
| `ai` | `openai` / `anthropic` | `chat` | browser-ext: chatgpt.com / claude.ai |
| `refactor` | — | — | replace > threshold с inserted >= replaced × 0.5 |
| `reading` | — | — | focus_ms без changes |

## VS Code / Cursor extension

Файл: `ide-vscode/src/extension.ts`. Двойная heuristic:

1. **Single big insert** (`inserted >= 80 chars` за один `contentChange`) →
   `ai_inline`. Это классический «accept Copilot suggestion» или paste из
   chat'а.
2. **Burst** (несколько `contentChange` в течение <100ms, суммарно >= 80) →
   тоже `ai_inline`. Inline streaming completion от Copilot/Cursor пишет
   токенами по 1-3 chars; обычный typing — interval >150ms между keystrokes.

`ai_provider` определяется через `vscode.env.appName`:
- `"Cursor"` → `cursor`
- иначе → `copilot` (default для VS Code inline-suggestions)

`app_bundle` тоже привязан к IDE: `com.todesktop.230313mzl4w4u92` (Cursor) или
`com.microsoft.VSCode`.

### Известные ограничения

- Real Copilot Accept event требует `proposed API`
  (`commands.onWillExecuteCommand`) — пока недоступен в stable VS Code.
  Burst-heuristic покрывает 85%+ случаев, остальное идёт как paste.
- Cursor's `aichat.acceptDiff` нельзя intercept через public API — same heuristic.
- Если Copilot выполняет очень медленную stream'инг (>100ms между chars), burst
  detection пропустит и attributирует как typing. Можно tune `BURST_WINDOW_MS`
  если pattern меняется.

## Claude Code hook

Самый сильный сигнал: Claude Code сам триггерит hook на каждом
`Edit`/`Write`/`MultiEdit` — нулевая ambiguity.

### Установка

1. Build hook binary:

   ```sh
   cd backend && go build -o ~/.local/bin/eop-hook ./cmd/eop-hook
   ```

   (или через `go install ./cmd/eop-hook` чтобы попало в `$GOBIN`).

2. Получи EOP token (один раз):

   ```sh
   curl -X POST https://eop.rysdavletov.org/api/v1/auth/dev-token | jq -r .token
   ```

   Сохрани в `~/.zshrc` / `~/.bashrc`:

   ```sh
   export EOP_TOKEN="<token>"
   export EOP_BACKEND="https://eop.rysdavletov.org/api"
   ```

3. Подключи hook в Claude Code settings (`~/.claude/settings.json` для
   user-wide или `.claude/settings.json` в репо для project-only):

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

4. Перезапусти Claude Code session. Каждый Edit/Write теперь шлёт event с
   `category=ai`, `ai_provider=claude-code`, `ai_channel=agent`.

### Что считается

- `Write` — вся длина `content` в `chars_in`, `\n` count → `lines_added`.
- `Edit` — длина `new_string` в chars, `\n` в `new_string` → `lines_added`,
  `\n` в `old_string` → `lines_removed`.
- `MultiEdit` — суммарно по всем edits.
- `Read` / `Bash` / `Grep` — игнорируются (нет attribution-смысла).

### Privacy

Hook шлёт только counts (chars/lines/lang), **не контент** файлов. Полная
схема в `validEvent()` (`backend/internal/ingest/handler.go`) гарантирует
отсутствие text content в payload.

### Failure-mode

Если `EOP_TOKEN` не задан — hook молча выходит (exit 0). Network error —
печатается в stderr, но не блокирует Claude Code workflow. Это by design: hook
не должен ломать tool-цикл при недоступности backend'а.

## Browser extension

Файл: `browser-extension/src/background.ts`. Без изменений в v2 — отслеживает
focus на AI-доменах, эмитит `category=ai`, `ai_provider` из mapping
`ai-domains.ts`, `ai_channel=chat` (не inline).

## Tauri agent

Файл: `agent/src-tauri/src/core/`. Эмитит focus events на native applications
(VS Code, Cursor, Claude Desktop). Для AI-attribution полагается на
parallel-running VS Code extension / browser extension. v2 не меняет.

## Roadmap дальше

- **Apply** для Cursor: добавить wraper-команду `eop.captureCursorApply` через
  registerCommand + keybinding override `cmd+enter` → executes Cursor's apply
  + emits explicit `ai` event. Точная attribution для Cursor users.
- **Copilot Accept**: дождаться stable `commands.onWillExecuteCommand` или
  proposed `vscode.copilot` API. На текущем stable burst-heuristic — best.
- **JetBrains plugin**: аналогичный VS Code ext но для IntelliJ family
  (PyCharm/WebStorm/etc). Большой sprint, отложен.
