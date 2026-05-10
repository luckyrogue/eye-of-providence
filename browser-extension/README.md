# Eye of Providence — browser extension

MV3 extension: tracks focus time on AI sites + clipboard-paste events from
ChatGPT/Claude/etc, шлёт metadata-only events на EoP backend.

## Local dev

```sh
pnpm install
pnpm -F @eop/browser-extension dev      # watch-build в dist/
pnpm -F @eop/browser-extension build    # production build в dist/
```

## Установка в Chrome / Edge / Brave

1. Build: `pnpm -F @eop/browser-extension build`
2. Открыть `chrome://extensions`
3. Включить **Developer mode** (toggle справа сверху)
4. Нажать **"Load unpacked"**
5. **Выбрать `browser-extension/dist/`** ← важно: НЕ корневую папку расширения

> **Не загружай корневой `browser-extension/`** — там исходный `manifest.json`
> который ссылается на `src/*.ts` (TypeScript). Chrome MV3 принимает только
> скомпилированный `.js`. Если grub'нул error
> *"content scripts can only be loaded from supported JavaScript files"* —
> ты загрузил исходник, а не `dist/`.

## Установка в Firefox (about:debugging)

1. Build
2. `about:debugging#/runtime/this-firefox`
3. **"Load Temporary Add-on"** → выбрать `browser-extension/dist/manifest.json`

Firefox-специфичные ограничения (нет HMR'а, перезагружать вручную после build).

## Архитектура

```
src/background.ts   — MV3 service worker (focus tracker, retry queue, push handler)
src/content.ts      — injected на AI-доменах, ловит clipboard paste events
src/popup.tsx       — UI (React) при клике иконки расширения
src/api.ts          — fetch helpers + chrome.storage.local config
src/ai-domains.ts   — provider-mapping (host → ai_provider, ai_channel)
manifest.json       — crxjs source (НЕ финальный — после build → dist/manifest.json)
```

`crxjs/vite-plugin` читает корневой `manifest.json`, transpiles `*.ts` → `*.js`,
переписывает пути и кладёт результат в `dist/`. Финальный extension =
`dist/`.

## Privacy

- Только metadata: host, duration, paste-size. **Не** content страниц или paste'нутый текст.
- Backend URL и token хранятся в `chrome.storage.local`.
- HMAC-подпись делается на backend webhooks-side (см. `internal/webhooks`).
