# Code attribution v2

Goal: 90%+ accuracy on «AI vs human» classification by hunks/strokes. The current
v1 algorithm (paste size + AI domains) achieves ~70%. v2 adds 3 signal
sources:

1. **VS Code/Cursor inline accept** — heuristic burst-detection
2. **Claude Code agent edits** — via [Hooks API](https://docs.claude.com/en/docs/claude-code/hooks)
3. **Browser AI domains** — claude.ai, chatgpt.com, perplexity, etc. (v1, unchanged)

## Categories and useful attribution

Two category layers:

- **Raw `events.category`** (ingest): `idle`, `manual`, `ai`, `reading`, `refactor`, `other`.
- **Derived `attribution_events.category`** (worker): `typed`, `pasted_ai`, `pasted_other`,
  `ai_inline`, `ai_agent`, `refactor`, `unknown` — see
  [`002_attribution_events.up.sql`](../backend/internal/migrate/sql/clickhouse/002_attribution_events.up.sql).

The table below maps **raw** events to signals (before worker Phase B):

| raw `category` | `ai_provider` | `ai_channel` | Signal |
|---|---|---|---|
| `manual` | — | — | typing / paste without AI tag |
| `ai` | `copilot` | `inline` | VS Code burst or single paste >= threshold |
| `ai` | `cursor` | `inline` | Cursor (`vscode.env.appName === "Cursor"`) + same heuristic |
| `ai` | `claude-code` | `agent` | `eop-hook` on PostToolUse `Edit\|Write\|MultiEdit` |
| `ai` | `openai` / `anthropic` | `chat` | browser-ext: chatgpt.com / claude.ai |
| `refactor` | — | — | structural change in IDE |
| `reading` | — | — | focus without edits |

## VS Code / Cursor extension

File: `ide-vscode/src/extension.ts`. Dual heuristic:

1. **Single big insert** (`inserted >= 80 chars` in one `contentChange`) →
   `ai_inline`. Classic «accept Copilot suggestion» or paste from
   chat.
2. **Burst** (several `contentChange` within <100ms, total >= 80) →
   also `ai_inline`. Inline streaming completion from Copilot/Cursor writes
   1–3 chars per token; normal typing — interval >150ms between keystrokes.

`ai_provider` is determined via `vscode.env.appName`:
- `"Cursor"` → `cursor`
- otherwise → `copilot` (default for VS Code inline suggestions)

`app_bundle` is tied to the IDE: `com.todesktop.230313mzl4w4u92` (Cursor) or
`com.microsoft.VSCode`.

### Known limitations

- Real Copilot Accept event requires `proposed API`
  (`commands.onWillExecuteCommand`) — not yet available in stable VS Code.
  Burst-heuristic covers 85%+ of cases; the rest goes as paste.
- Cursor's `aichat.acceptDiff` cannot be intercepted via public API — same heuristic.
- If Copilot streams very slowly (>100ms between chars), burst
  detection may miss and attribute as typing. Tune `BURST_WINDOW_MS`
  if the pattern changes.

## Claude Code hook

Strongest signal: Claude Code triggers the hook on every
`Edit`/`Write`/`MultiEdit` — zero ambiguity.

### Installation

1. Build hook binary:

   ```sh
   cd backend && go build -o ~/.local/bin/eop-hook ./cmd/eop-hook
   ```

   (or `go install ./cmd/eop-hook` to place it in `$GOBIN`).

2. Get EOP token (**local dev only**; in production `dev-token` is disabled):

   ```sh
   # backend with EOP_ENABLE_DEV_TOKEN=true (default in development)
   curl -X POST http://localhost:8080/v1/auth/dev-token | jq -r .token
   ```

   Save in `~/.zshrc` / `~/.bashrc`:

   ```sh
   export EOP_TOKEN="<token>"
   export EOP_BACKEND="http://localhost:8080"   # or your self-host URL
   ```

   In production use an API token from the dashboard (`eop_<…>`) or device pairing.

3. Connect the hook in Claude Code settings (`~/.claude/settings.json` for
   user-wide or `.claude/settings.json` in the repo for project-only):

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

4. Restart the Claude Code session. Each Edit/Write now sends an event with
   `category=ai`, `ai_provider=claude-code`, `ai_channel=agent`.

### What is counted

- `Write` — full length of `content` in `chars_in`, `\n` count → `lines_added`.
- `Edit` — length of `new_string` in chars, `\n` in `new_string` → `lines_added`,
  `\n` in `old_string` → `lines_removed`.
- `MultiEdit` — sum across all edits.
- `Read` / `Bash` / `Grep` — ignored (no attribution meaning).

### Privacy

The hook sends only counts (chars/lines/lang), **not file content**. Validation
on ingest — `domain.ValidEvent()` in
[`backend/internal/ingest/domain/event.go`](../backend/internal/ingest/domain/event.go);
no text fields in the wire format.

### Failure mode

If `EOP_TOKEN` is unset — hook exits silently (exit 0). Network error —
printed to stderr but does not block Claude Code workflow. By design: the hook
must not break the tool loop when the backend is unavailable.

## Browser extension

File: `browser-extension/src/entrypoints/background.ts`. Unchanged in v2 — tracks
focus on AI domains, emits `category=ai`, `ai_provider` from mapping
`ai-domains.ts`, `ai_channel=chat` (not inline).

## Tauri agent

File: `agent/src-tauri/src/core/`. Emits focus events on native applications
(VS Code, Cursor, Claude Desktop). For AI attribution it relies on
parallel-running VS Code extension / browser extension. v2 unchanged.

## Roadmap

- **Apply** for Cursor: add wrapper command `eop.captureCursorApply` via
  registerCommand + keybinding override `cmd+enter` → runs Cursor's apply
  + emits explicit `ai` event. Precise attribution for Cursor users.
- **Copilot Accept**: wait for stable `commands.onWillExecuteCommand` or
  proposed `vscode.copilot` API. On current stable burst-heuristic is best.
- **JetBrains plugin**: VS Code ext analogue for IntelliJ family
  (PyCharm/WebStorm/etc). Large sprint, deferred.
