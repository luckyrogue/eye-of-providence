# Alpha — client installation

Document for alpha participants (v0.1.x). Contact the maintainer if
you need access to artifacts (GH Release is still draft).

## Pre-requisites

1. Dashboard account (https://eop.rysdavletov.org). Register via
   email/passcode.
2. One supported client installed (see below).
3. A 6-character pairing code from the dashboard (`Settings → Devices →
   Claim device`) — valid for 10 minutes; generate a new one after expiry.

## 1. Browser extension (Chrome / Edge / Brave)

**Source:** `eop-browser-extension.zip` from GH Release (draft).

```bash
# Unpack
unzip eop-browser-extension.zip -d eop-browser-extension/
```

1. Open `chrome://extensions/`.
2. Enable **Developer mode** (top right).
3. **Load unpacked** → select the `eop-browser-extension/` folder.
4. Pin the icon in the toolbar.
5. Click the icon → **Start pairing**.
6. Dashboard `Settings → Devices` opens; the code is copied to the clipboard.
7. Paste the code and click `Claim`. The extension shows `paired`.

**Verification:** open ChatGPT / Claude / Gemini, type something. After ~30s
click the popup — `Pending: 1+` should appear. Events reach the dashboard in 1–2 minutes.

## 2. Agent (macOS / Windows / Linux)

**Source:** `.dmg` / `.msi` / `.AppImage` from GH Release (draft).

### macOS

```bash
# DMG → drag .app to /Applications
open ~/Downloads/eop-agent_*.dmg
```

First launch:

1. **System Settings → Privacy & Security → Accessibility** → enable EoP
   (required for active window tracking).
2. Open EoP → **Settings** tab → **Start pairing**.
3. Open the dashboard link from the card and enter the 6-character code.
4. After `paired`, the agent starts writing events.

Logs: `Settings → Open logs folder` → `~/Library/Application Support/com.eop.agent/logs/`.

### Windows

1. Run `.msi`, accept the SmartScreen warning (unsigned build).
2. After install it runs from the tray.
3. Open from tray → **Settings** tab → **Start pairing**.

### Linux

```bash
chmod +x eop-agent_*.AppImage
./eop-agent_*.AppImage
```

`.AppImage` is self-contained, no install required. For autostart see
**Settings → Autostart**.

## 3. VS Code extension

**Source:** `eop-vscode.vsix` from GH Release.

```bash
code --install-extension eop-vscode.vsix
```

or via VS Code UI: **Extensions** → `…` → **Install from VSIX…**.

After install:

1. Welcome window opens.
2. Cmd/Ctrl+Shift+P → **EoP: Pair this editor**.
3. Dashboard opens in the browser; code is copied to the clipboard.
4. Paste the code in Devices → Claim.
5. Status bar shows `$(eye) EoP idle` — extension is ready.

**What we track:** manual character input in the editor (`chars_in`),
keystroke counts, AI-pasted lines (diffs from the buffer).

**What we do not track:** file contents, file names, or codebase.

## Troubleshooting

| Symptom | What to do |
| --- | --- |
| `auth required` in popup/status bar | Token revoked. Settings → Devices → revoke + create new. |
| `Pending: N+` does not decrease | Backend unavailable or token expired. Check `Open logs` or dashboard health. |
| Pairing code expired | Generate a new one — `Claim device` button in the dashboard. |
| macOS: no events | Check Accessibility permission. Without it the agent sees bundle.id only, not keystrokes. |
| Linux/AppImage won't open | Install `libwebkit2gtk-4.1`, `libgtk-3` (see release notes). |

## Privacy

- No URLs, page titles, file names, or code — only
  `app_bundle` (e.g. `chat.openai.com`), category (`ai` / `manual`),
  and counters (`chars_in`, `duration_ms`).
- Pairing token is stored in the OS keyring (macOS Keychain / Windows Credential
  Manager / GNOME keyring), not in plaintext.
- Full event schema: [/docs/data-model](/docs/data-model).
- Threat model: [/security](/security).

## Submit feedback

GitHub Issues → tag `alpha-feedback`. Include:

- Client + version (see `Settings → About` or `vsce show`).
- OS + version.
- Steps to reproduce.
- Log contents (without personally identifiable strings).
