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

> **⚠ Note про unsigned binaries.** Alpha-installer не подписан Apple
> Developer ID / Windows EV cert (см. [FAQ](#почему-installer-не-подписан)
> ниже). При первом запуске ОС покажет warning — это **ожидаемо**, не
> признак compromise. Workaround ниже.

### macOS

```bash
# DMG → drag .app в /Applications
open ~/Downloads/Eye.of.Providence_*_*.dmg
```

**Если при запуске видишь `"Eye of Providence" cannot be opened because Apple cannot check it for malicious software`:**

- **macOS 14 (Sonoma) и ниже:** правый клик по `Eye of Providence.app`
  в Finder → **Open** → во втором модале снова **Open**. Один раз.
- **macOS 15+ (Sequoia):** правый-клик bypass убран. Идём в **System
  Settings → Privacy & Security**, прокрутить вниз до раздела «Security»
  → рядом со строкой "Eye of Providence was blocked..." нажать
  **Open Anyway** → подтвердить Touch ID / password.
- **Alternative (CLI):** `xattr -dr com.apple.quarantine /Applications/Eye\ of\ Providence.app`
  — снимает quarantine флаг, OS перестаёт спрашивать.

Первый запуск после bypass:

1. **System Settings → Privacy & Security → Accessibility** → включить
   "Eye of Providence" (нужно для расширенных сигналов attribution;
   базовые keystroke counts работают и без него).
2. Открыть EoP → tab **Settings** → **Start pairing**.
3. Открой dashboard ссылкой из карточки, введи 6-символьный код.
4. После `paired` агент начнёт писать события.

Логи: `Settings → Open logs folder` → `~/Library/Application Support/com.eop.agent/logs/`.

### Windows

```cmd
:: Run installer (или дабл-клик в Explorer)
msiexec /i Eye.of.Providence_*_x64_en-US.msi
```

**Если SmartScreen показывает `Windows protected your PC. Microsoft
Defender SmartScreen prevented an unrecognized app from starting`:**

1. Нажать **More info** (мелкая ссылка под title'ом, легко проскочить).
2. Появится кнопка **Run anyway** → нажать.
3. Установщик запустится как обычно.

**Если в корпоративной среде с AppLocker / WDAC:** installer
заблокирован системно, bypass отсутствует. Попроси админа добавить
hash в allow-list — выложен в `Eye.of.Providence_*_x64-setup.exe.sig`
(SHA-256). Или жди signed build (см. [FAQ](#почему-installer-не-подписан)).

После установки:

1. Запустится из system tray (иконка глаза справа внизу).
2. Открыть из tray → tab **Settings** → **Start pairing**.
3. Дальше как macOS.

### Linux

```bash
# AppImage — portable, ничего не ставится в систему
chmod +x Eye.of.Providence_*_amd64.AppImage
./Eye.of.Providence_*_amd64.AppImage

# Или через .deb (Ubuntu / Debian / Mint)
sudo dpkg -i Eye.of.Providence_*_amd64.deb

# Или через .rpm (Fedora / openSUSE)
sudo rpm -i Eye.of.Providence-*_x86_64.rpm
```

`.AppImage` self-contained, требует только `libwebkit2gtk-4.1` +
`libgtk-3` в системе (см. **Trouble-shooting** если падает).

> **⚠ Linux parity gap.** Сейчас Linux-агент собирает только foreground
> app + idle (нет keystroke counts / clipboard digest как на macOS).
> Tracking в `docs/tech-debt.md` cluster C6. Полный паритет → v0.2.

Autostart: **Settings → Autostart** (systemd user service).

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

## FAQ

### Почему installer не подписан?

В alpha-фазе нет Apple Developer ID ($99/год) и Windows EV cert
($300+/год). Это сознательное решение — пока user base = знакомые +
ранние участники, $400/год cash-burn до первой выручки не оправдан.

**Что мы делаем вместо подписи:**

1. **Прозрачная supply chain.** Docker image (`ghcr.io/luckyrogue/eop`)
   подписан через Cosign keyless (Sigstore OIDC) с SLSA Build L3
   attestation. Можно верифицировать `cosign verify` (см.
   [.github/SECURITY.md](../.github/SECURITY.md)). Agent installer
   собирается из того же git-commit'а — если backend image checks out,
   агент тоже.
2. **Воспроизводимая сборка.** Tag + commit указаны в каждом GH Release.
   Можно склонировать репо и собрать installer самостоятельно
   (`pnpm -F @eop/agent tauri build`).
3. **Public source.** Весь код agent'а в репо. Поведение можно audit'ить
   без рантайма (или после установки — `Settings → Open logs folder`).

**Когда появится подпись:** при переходе alpha → beta (см. cluster C1
в [tech-debt.md](tech-debt.md)). Apple cert первым (Sequoia-friction
выше всего), Windows EV — когда appear первые paying customers или
enterprise leads.

### Зачем macOS Accessibility permission?

Без него агент видит:
- Текущее foreground-приложение (`com.apple.dt.Xcode`, `chat.openai.com`)
- Idle/active state (через `CGEventSourceSecondsSinceLastEventType` —
  публичный API, **разрешения не требует**)
- Keystroke counts (через `CGEventSourceCounterForEventType` — тоже
  публичный API без разрешений)
- Clipboard fingerprint (через `NSPasteboard.changeCount` + sha256)

С Accessibility unlock'аются расширенные сигналы attribution: focused
text-field detection, IDE-specific paste origin tracking. На текущей
alpha они НЕ обязательны — базовая статистика собирается без них.

### Сколько данных утекает наружу?

Только metadata: имена приложений, длительности, hex-хеши clipboard,
provider/channel метки для AI событий. **Никогда:** содержимое файлов,
prompts, ответы AI, raw keystrokes, текст clipboard'а, имена файлов,
URL-параметры. Полная карта потоков — [`docs/privacy.md`](privacy.md).

### Backend self-hosted? Что отличается?

Если ты сам поднимаешь backend (`docker-compose.full.yml`) — данные не
покидают твою инфраструктуру вообще, кроме:
- Gemini API при `EOP_GEMINI_API_KEY=<set>` (агрегаты для AI-отчётов)
- Resend при `EOP_RESEND_API_KEY=<set>` (transactional email)
- GitHub OAuth при `EOP_GITHUB_CLIENT_ID=<set>`

Все три опциональны. См. [`docs/self-hosting.md`](self-hosting.md).

## Опубликовать feedback

GitHub Issues → tag `alpha-feedback`. Опиши:

- Клиент + версия (см. `Settings → About` или `vsce show`).
- ОС + версия.
- Шаги повтора.
- Содержимое логов (без personally-identifiable строк).
