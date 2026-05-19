# Alpha — установка клиентов

Документ для участников alpha (v0.1.x). Контактируйте maintainer'а, если
нужен доступ к артефактам (GH Release пока draft).

## Pre-requisites

1. Аккаунт на dashboard'е (https://eop.rysdavletov.org). Регистрация через
   email/passcode.
2. Один из поддерживаемых клиентов установлен (см. ниже).
3. Готов 6-символьный pairing-код из dashboard'а (`Settings → Devices →
   Claim device`) — он живёт 10 минут, после нужно сгенерировать новый.

## 1. Browser extension (Chrome / Edge / Brave)

**Источник:** `eop-browser-extension.zip` из GH Release (draft).

```bash
# Распаковать
unzip eop-browser-extension.zip -d eop-browser-extension/
```

1. Открыть `chrome://extensions/`.
2. Включить **Developer mode** (правый верхний угол).
3. **Load unpacked** → выбрать папку `eop-browser-extension/`.
4. Закрепить иконку в toolbar.
5. Кликнуть на иконку → **Start pairing**.
6. Откроется `Settings → Devices` в dashboard'е, код скопируется в clipboard.
7. Вставить код, нажать `Claim`. Расширение покажет `paired`.

**Проверка:** открой ChatGPT / Claude / Gemini, попечатай. Через ~30s
кликни на popup — `Pending: 1+` должно появиться. Через 1-2 минуты события
доедут в dashboard.

## 2. Agent (macOS / Windows / Linux)

**Источник:** `.dmg` / `.msi` / `.AppImage` из GH Release (draft).

### macOS

```bash
# DMG → drag .app в /Applications
open ~/Downloads/eop-agent_*.dmg
```

Первый запуск:

1. **System Settings → Privacy & Security → Accessibility** → включить EoP
   (нужно для отслеживания active window).
2. Открыть EoP → tab **Settings** → **Start pairing**.
3. Открой dashboard ссылкой из карточки, введи 6-символьный код.
4. После `paired` агент начнёт писать события.

Логи: `Settings → Open logs folder` → `~/Library/Application Support/com.eop.agent/logs/`.

### Windows

1. Запустить `.msi`, согласиться на SmartScreen warning (unsigned build).
2. После установки запустится из tray.
3. Открыть из tray → tab **Settings** → **Start pairing**.

### Linux

```bash
chmod +x eop-agent_*.AppImage
./eop-agent_*.AppImage
```

`.AppImage` self-contained, не требует установки. Для autostart смотри
**Settings → Autostart**.

## 3. VS Code extension

**Источник:** `eop-vscode.vsix` из GH Release.

```bash
code --install-extension eop-vscode.vsix
```

или через VS Code UI: **Extensions** → `…` → **Install from VSIX…**.

После установки:

1. Откроется welcome-окно.
2. Cmd/Ctrl+Shift+P → **EoP: Pair this editor**.
3. Откроется dashboard в браузере, код скопируется в clipboard.
4. Вставить код в Devices → Claim.
5. В status bar внизу появится `$(eye) EoP idle` — extension готов.

**Что трекаем:** ручной ввод символов в редакторе (расчёт `chars_in`),
keystroke counts, AI-pasted lines (диффы из buffer'а).

**Не трекаем:** содержимое файлов, имена файлов, кодовая база.

## Trouble-shooting

| Симптом | Что делать |
| --- | --- |
| `auth required` в popup/status-bar | Token revoked. Settings → Devices → revoke + создать новый. |
| `Pending: N+` не уменьшается | Backend недоступен или token истёк. Проверь `Open logs` или dashboard health. |
| Pairing code expired | Сгенерировать новый — кнопка `Claim device` в dashboard'е. |
| macOS: события не идут | Проверь Accessibility permission. Без него agent видит только bundle.id, не keystrokes. |
| Linux/AppImage не открывается | Установить `libwebkit2gtk-4.1`, `libgtk-3` (см. release notes). |

## Privacy

- Никаких URL'ов, заголовков страниц, имён файлов или кода — только
  `app_bundle` (e.g. `chat.openai.com`), категория (`ai` / `manual`),
  и счётчики (`chars_in`, `duration_ms`).
- Pairing-token хранится в OS keyring (macOS Keychain / Windows Credential
  Manager / GNOME keyring), не в plaintext.
- Полная схема событий: [docs/data-model.md](data-model.md).
- Threat model: [docs/threat-model.md](threat-model.md).

## Опубликовать feedback

GitHub Issues → tag `beta-feedback`. Опиши:

- Клиент + версия (см. `Settings → About` или `vsce show`).
- ОС + версия.
- Шаги повтора.
- Содержимое логов (без personally-identifiable строк).
