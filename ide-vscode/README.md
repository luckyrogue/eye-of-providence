# Eye of Providence — VS Code / Cursor extension

Tracks how much code in your editor is typed manually versus accepted from
Copilot / Cursor inline completions, then ships per-language attribution
statistics to your Eye of Providence dashboard.

## What it captures

- Manual typing — small, slow insertions.
- AI inline accept — bursts of small inserts or single big inserts (over
  `eop.pasteThresholdChars`, default 80).
- Refactor — large replace operations.
- Idle / focus duration per language.

No source code is ever sent to the backend, only aggregated counters.

## Pair this editor

1. Run command **Eye of Providence: Pair this editor** (Cmd/Ctrl+Shift+P).
2. The extension opens the dashboard and shows a 6-character code.
3. In the dashboard, go to **Settings → Connected devices** and enter the code.
4. The extension finishes pairing automatically — status bar switches to
   `EoP idle`.

The pairing token is stored in VS Code SecretStorage (system keychain),
never on disk in plain text.

## Status bar

- `$(eye) EoP idle` — extension active, queue empty.
- `$(sync~spin) EoP sending` — flush in progress.
- `$(warning) EoP auth required` — token revoked or expired; click to re-pair.
- `$(circle-slash) EoP paused` — tracking paused.

Click the status bar item to open the dashboard.

## Commands

| Command | Description |
| --- | --- |
| `Eye of Providence: Pair this editor` | Start pairing flow. |
| `Eye of Providence: Sign out` | Forget the stored token. |
| `Eye of Providence: Flush attribution events` | Push current buffer immediately. |
| `Eye of Providence: Open dashboard` | Open the dashboard in your browser. |
| `Eye of Providence: Show log` | Reveal the output channel. |

## Settings

- `eop.backendUrl` — base URL of your EoP backend (default
  `https://eop.rysdavletov.org/api`).
- `eop.flushIntervalSec` — how often to flush events (default 30s).
- `eop.pasteThresholdChars` — minimum chars in a single insert to be treated
  as AI / paste (default 80).

## Status: closed beta

Distributed as a `.vsix` for invited testers. Install via
`code --install-extension eop-vscode-0.0.1.vsix` or **Extensions →
Install from VSIX…**.
