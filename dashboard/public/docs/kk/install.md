# Alpha — клиенттерді орнату

Alpha қатысушылары үшін құжат (v0.1.x). Артефактілерге қол жеткізу керек болса
maintainer-мен байланысыңыз (GH Release әзірге draft).

## Алдын ала талаптар

1. Dashboard аккаунты (https://eop.rysdavletov.org). Email/passcode арқылы тіркелу.
2. Қолдау көрсетілетін клиенттердің бірі орнатылған (төменге қараңыз).
3. Dashboard-тан 6 таңбалы pairing коды (`Settings → Devices →
   Claim device`) — 10 минут жарамды; мерзімі өткеннен кейін жаңасын генерациялаңыз.

## 1. Browser extension (Chrome / Edge / Brave)

**Көз:** GH Release-тен `eop-browser-extension.zip` (draft).

```bash
# Распаковка
unzip eop-browser-extension.zip -d eop-browser-extension/
```

1. `chrome://extensions/` ашыңыз.
2. **Developer mode** қосыңыз (оң жақ жоғары бұрыш).
3. **Load unpacked** → `eop-browser-extension/` қалтасын таңдаңыз.
4. Toolbar-ға иконканы бекітіңіз.
5. Иконкаға басыңыз → **Start pairing**.
6. Dashboard `Settings → Devices` ашылады, код clipboard-қа көшіріледі.
7. Кодты қойып `Claim` басыңыз. Extension `paired` көрсетеді.

**Тексеру:** ChatGPT / Claude / Gemini ашып, біраз теріңіз. ~30 сек кейін
popup басыңыз — `Pending: 1+` пайда болуы керек. 1–2 минутта оқиғалар dashboard-қа жетеді.

## 2. Agent (macOS / Windows / Linux)

**Көз:** GH Release-тен `.dmg` / `.msi` / `.AppImage` (draft).

### macOS

```bash
# DMG → .app файлын /Applications ішіне сүйреу
open ~/Downloads/eop-agent_*.dmg
```

Алғашқы іске қосу:

1. **System Settings → Privacy & Security → Accessibility** → EoP қосу
   (active window бақылау үшін қажет).
2. EoP ашу → **Settings** tab → **Start pairing**.
3. Карточкадағы dashboard сілтемесін ашып 6 таңбалы кодты енгізіңіз.
4. `paired` кейін agent оқиғалар жаза бастайды.

Логтар: `Settings → Open logs folder` → `~/Library/Application Support/com.eop.agent/logs/`.

### Windows

1. `.msi` іске қосу, SmartScreen ескертуін қабылдау (unsigned build).
2. Орнатудан кейін tray-дан іске қосылады.
3. Tray → **Settings** tab → **Start pairing**.

### Linux

```bash
chmod +x eop-agent_*.AppImage
./eop-agent_*.AppImage
```

`.AppImage` self-contained, орнату қажет емес. Autostart үшін
**Settings → Autostart** қараңыз.

## 3. VS Code extension

**Көз:** GH Release-тен `eop-vscode.vsix`.

```bash
code --install-extension eop-vscode.vsix
```

немесе VS Code UI: **Extensions** → `…` → **Install from VSIX…**.

Орнатудан кейін:

1. Welcome терезесі ашылады.
2. Cmd/Ctrl+Shift+P → **EoP: Pair this editor**.
3. Dashboard браузерде ашылады, код clipboard-қа көшіріледі.
4. Devices → Claim ішінде кодты қойыңыз.
5. Status bar `$(eye) EoP idle` көрсетеді — extension дайын.

**Не бақылаймыз:** редактордағы қолмен енгізу (`chars_in`),
keystroke counts, AI-pasted жолдар (buffer diff).

**Не бақыламаймыз:** файл мазмұны, файл атаулары, codebase.

## Troubleshooting

| Белгі | Не істеу керек |
| --- | --- |
| popup/status-bar `auth required` | Token revoked. Settings → Devices → revoke + жаңасын жасау. |
| `Pending: N+` азаймайды | Backend қолжетімсіз немесе token мерзімі өткен. `Open logs` немесе dashboard health тексеріңіз. |
| Pairing code expired | Жаңасын генерациялау — dashboard `Claim device`. |
| macOS: оқиғалар жоқ | Accessibility рұқсатын тексеріңіз. Онсыз agent тек bundle.id көреді, keystrokes емес. |
| Linux/AppImage ашылмайды | `libwebkit2gtk-4.1`, `libgtk-3` орнатыңыз (release notes). |

## Privacy

- URL, бет тақырыптары, файл атаулары немесе код жоқ — тек
  `app_bundle` (мыс. `chat.openai.com`), category (`ai` / `manual`),
  және санағыштар (`chars_in`, `duration_ms`).
- Pairing token OS keyring-де (macOS Keychain / Windows Credential
  Manager / GNOME keyring), plaintext емес.
- Толық event схемасы: [/docs/data-model](/docs/data-model).
- Threat model: [/security](/security).

## Feedback жіберу

GitHub Issues → `alpha-feedback` тегі. Көрсетіңіз:

- Клиент + нұсқа (`Settings → About` немесе `vsce show`).
- ОЖ + нұсқа.
- Қайталау қадамдары.
- Лог мазмұны (жеке сәйкестендірілетін жолдарсыз).
