# Agents publishing checklist

3 агента, 3 marketplace'а, 3 разных процесса. Эта дока — что нужно сделать
ОДИН РАЗ для перехода с локальных билдов на public-distribution.

---

## 1. Tauri desktop (macOS + Windows)

### macOS — Code-signing + notarization

**Зачем:** Без подписи macOS Gatekeeper показывает "App is damaged". Без notarization —
"App can't be opened, developer cannot be verified".

**Что нужно:**

1. **Apple Developer Program** ($99/год) — https://developer.apple.com/programs/
2. **Developer ID Application certificate** в Keychain Access (через Apple Developer
   portal → Certificates → "Developer ID Application")
3. **App-specific password** для Apple ID — https://account.apple.com/account/manage
   → Sign-in and Security → App-Specific Passwords

**Setup в репо (GitHub Actions secrets):**

```bash
# Экспортируй cert как .p12 (Keychain → Right-click → Export → .p12 + пароль)
base64 -i certificate.p12 -o cert.b64
# Содержимое cert.b64 → secret APPLE_CERTIFICATE (base64-encoded)
```

| Secret | Что |
|---|---|
| `APPLE_CERTIFICATE` | base64 .p12 |
| `APPLE_CERTIFICATE_PASSWORD` | пароль от .p12 |
| `APPLE_SIGNING_IDENTITY` | строка типа `Developer ID Application: Your Name (TEAMID)` |
| `APPLE_ID` | твой Apple ID email |
| `APPLE_PASSWORD` | app-specific password (НЕ пароль от Apple ID) |
| `APPLE_TEAM_ID` | 10-char Team ID из Apple Developer portal |

`tauri-apps/tauri-action@v0` (в `.github/workflows/release.yml`) автоматически
подхватит эти secrets и сделает sign + notarize.

### Windows — Code-signing

**Зачем:** Без signature SmartScreen блокирует skip-warning. С EV (Extended Validation)
cert — instant trust. Без — придётся "build reputation" через первые ~3000 installs.

**Варианты:**

- **Microsoft Store** ($19 one-time для individual) — auto-sign + distribution. Удобно
  но требует review.
- **Standard Code-Signing Cert** ($150-400/год от GoDaddy/DigiCert) — sign + распространять
  где угодно. Reputation building first time.
- **EV Cert** ($300-600/год + USB hardware token) — instant trust. Дорого, но не нужен
  build reputation.

**Setup secrets:**

| Secret | Что |
|---|---|
| `WINDOWS_CERTIFICATE` | base64 .pfx |
| `WINDOWS_CERTIFICATE_PASSWORD` | пароль от .pfx |

**Self-hosted updater (опц.):** Tauri имеет встроенный updater plugin. Если хочешь
автообновления — добавить `tauri-plugin-updater` + endpoint в `tauri.conf.json`.
Endpoint = static JSON с metadata + download URL. Можно хостить на GitHub Releases
напрямую (`https://github.com/luckyrogue/eye-of-providence/releases/latest/download/latest.json`).

---

## 2. VS Code Marketplace

**Зачем:** Юзеры устанавливают через `code --install-extension eop-vscode` или из
UI marketplace.

### Setup (один раз)

1. Создать **Azure DevOps publisher**: https://aka.ms/vscode-create-publisher
2. Получить **Personal Access Token** на Azure DevOps:
   - User Settings → Personal Access Tokens
   - Scope: `Marketplace (manage)`
   - Expiration: 1 year max
3. Сохранить как secret `VSCE_TOKEN`

### Publish (per release)

```bash
cd ide-vscode
npm i -g @vscode/vsce

# Update version in package.json
vsce publish patch    # 0.0.1 → 0.0.2
# или
vsce publish 0.1.0
```

**Required в `package.json`:**

```json
{
  "publisher": "luckyrogue",
  "name": "eop-vscode",
  "displayName": "Eye of Providence",
  "description": "Track AI vs manual coding time. Privacy-first.",
  "icon": "icon.png",
  "repository": { "type": "git", "url": "https://github.com/luckyrogue/eye-of-providence" },
  "engines": { "vscode": "^1.85.0" },
  "categories": ["Other"]
}
```

### CI auto-publish (опц.)

Добавить step в `.github/workflows/release.yml`:

```yaml
- name: Publish to VS Code Marketplace
  if: startsWith(github.ref, 'refs/tags/')
  run: |
    cd ide-vscode
    npx vsce publish -p ${{ secrets.VSCE_TOKEN }}
```

---

## 3. Browser extension

### Chrome Web Store

**Setup:**

1. **Developer account**: $5 one-time fee — https://chrome.google.com/webstore/devconsole
2. **Verify identity** (Google Wallet) — занимает пару дней
3. Подготовить store assets:
   - **Promo tile**: 440×280 (small) + 920×680 (large) PNG
   - **Screenshots**: 1280×800 (минимум 1, максимум 5)
   - **Privacy policy URL**: должен быть public (можно `https://eop.rysdavletov.org/privacy`)

**Submit:**

1. Зайти в Developer Dashboard → Add new item
2. Upload `eop-browser-extension.zip` (артефакт из release.yml)
3. Заполнить листинг (название, описание, screenshots, privacy)
4. Submit for review — обычно 1-3 дня

**Update flow:**

Каждый новый release.yml run создаёт zip-артефакт. Manual upload через dashboard, либо
автоматизировать через [Chrome Web Store API](https://developer.chrome.com/docs/webstore/api):

```yaml
# В release.yml, опционально:
- uses: mnao305/chrome-extension-upload@v5
  with:
    file-path: browser-extension/eop-browser-extension.zip
    extension-id: ${{ vars.CHROME_EXTENSION_ID }}
    client-id: ${{ secrets.CHROME_CLIENT_ID }}
    client-secret: ${{ secrets.CHROME_CLIENT_SECRET }}
    refresh-token: ${{ secrets.CHROME_REFRESH_TOKEN }}
```

### Firefox AMO (addons.mozilla.org)

**Setup:**

1. Free account на https://addons.mozilla.org/developers/
2. **API credentials** (JWT): https://addons.mozilla.org/developers/addon/api/key/

**Submit:**

```bash
# Локально:
npm i -g web-ext
cd browser-extension/dist
web-ext sign \
  --api-key=$AMO_JWT_ISSUER \
  --api-secret=$AMO_JWT_SECRET \
  --channel=listed
# → выдаёт signed .xpi
```

Firefox review медленнее (1-3 недели первый раз, потом быстрее).

### Manifest V3 production checks

Перед submission — проверь:

- [ ] `manifest.json` `host_permissions` минимальные (только реально нужные домены)
- [ ] Нет `<all_urls>` если не критично
- [ ] `content_security_policy.extension_pages` без `unsafe-eval`/`unsafe-inline`
- [ ] Privacy policy URL валидный
- [ ] `permissions` justification документирован (Chrome спросит "why do you need
      `tabs`?", "why `storage`?")
- [ ] Background service worker не делает longlived listeners (MV3 убивает worker через
      30 сек idle)

---

## 4. GitHub Releases (текущий минимум)

`.github/workflows/release.yml` уже умеет:

- ✅ matrix builds для macOS / Windows / Linux Tauri
- ✅ VS Code .vsix package
- ✅ Browser extension .zip
- ✅ Tauri code-signing если secrets настроены (иначе unsigned warning)
- ✅ Tauri notarization для macOS если Apple secrets настроены
- ✅ Auto-create draft GH Release с прикреплёнными артефактами

**Чтобы выпустить:**

```bash
git tag v0.1.0
git push origin v0.1.0
# → release.yml запустится, создаст draft Release
# → ручной review → publish
```

Без single secret настроенного — pipeline всё равно проходит, артефакты unsigned (для
internal dogfooding).

---

## 5. Roadmap для агентов до prod-quality

### Сделано в этом репо

- ✅ MV3 service-worker state persistence (chrome.storage.session)
- ✅ Retry queue с exponential backoff (browser ext)
- ✅ Tauri pause/resume через PauseFlag (UI + tray)
- ✅ Tauri SQLite GC (события > 7 дней удаляются)
- ✅ VS Code multi-window dedup (focused-window-only tracking)
- ✅ Release pipeline для всех 3 агентов

### Следующие шаги (требуют user action)

- [ ] Apple Developer account → secrets для notarization
- [ ] Windows code-signing cert (Microsoft Store или standard cert)
- [ ] Chrome Web Store: $5 + identity verification + privacy policy
- [ ] VS Code Marketplace: Azure DevOps publisher + PAT
- [ ] Firefox AMO: free account + JWT
- [ ] Tauri auto-updater: endpoint JSON на GH Releases
- [ ] Privacy policy + Terms public-URL

### Defer до Q3-Q4

- [ ] Sentry или crash reporting в desktop agent
- [ ] Telemetry: how often agent crashes, ingest success rate
- [ ] Self-hosted enterprise distribution (.dmg для airgapped install)
- [ ] Linux .deb / Flatpak (сейчас только AppImage)
