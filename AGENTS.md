# Eye of Providence — frontend conventions

Единые правила для всех frontend-пакетов: `ui`, `dashboard`, `agent`, `browser-extension`.
Backend — см. `backend/AGENTS.md`. VS Code extension (`ide-vscode`) — Node-side, не входит в frontend standard.

## Layout

| Package           | Тип                   | FSD layers                                       |
| ----------------- | --------------------- | ------------------------------------------------ |
| ui                | shared library        | shared/{lib,ui}, widgets, features               |
| dashboard         | SPA (Vite + React)    | app, pages, widgets, features, entities, shared  |
| agent             | Tauri shell           | app, pages, shared                               |
| browser-extension | MV3 extension         | entrypoints, shared                              |

Агентские пакеты с малым кол-вом файлов используют усечённый набор слоёв.

## TypeScript

- Все `tsconfig.json` extends корневой `tsconfig.base.json`. Локальные overrides — только Tauri/chrome-specific (`useDefineForClassFields`, `allowImportingTsExtensions`, `types: ["chrome"]`).
- `strict` + `noUnusedLocals` + `noUnusedParameters` + `noFallthroughCasesInSwitch` везде.

## ESLint

- Все используют `@eop/ui/eslint.base.js` через `baseConfig()`.
- `pnpm -r lint` должен быть 0 ошибок. CI блокирует merge при failure.
- Guard-правило `no-restricted-syntax` запрещает голые `<button>` (использовать `Button`/`IconButton` из `@eop/ui`), `<input readOnly>` (использовать `SecretField`), нативные **`<input>`** и **`<select>`** (использовать `Input`/`Checkbox`/`SecretField` и `Select`/`SimpleSelect`/`SelectField` из `@eop/ui`).

## Prettier

- Корневой `.prettierrc.json` (`printWidth: 100`, `trailingComma: "all"`, `semi: true`, `singleQuote: false`).
- `pnpm format` форматирует, `pnpm format:check` валидирует. Backend/Go-файлы исключены через `.prettierignore`.

## Tailwind

- Все импортят `@eop/ui/tailwind.config.js` через фабрику `uiTailwindConfig({ content })`. Не дублировать preset руками.

## i18n

- Общий пакет **`@eop/i18n`**: `createI18n`, `SUPPORTED_LOCALES`, `LOCALE_LABELS`, `LOCALE_STORAGE_KEY` (`eop_locale`). Подключён в `dashboard`, `agent`, `browser-extension`.
- Локали приложений живут в `*/src/shared/i18n/locales/`; экземпляр i18next создаётся один раз и передаётся в дерево через `I18nextProvider` (`react-i18next`).

## UI-компоненты

См. `dashboard/AGENTS.md` → раздел "UI-компоненты" (правила shadcn-first, чек-лист, restricted patterns).
Применяется ко ВСЕМ frontend-пакетам.

## FSD-конвенции

См. `dashboard/AGENTS.md` → разделы про слои/сегменты.
Применяются ко всем пакетам, у которых есть FSD-структура.

## Команды

```bash
pnpm install
pnpm -r exec tsc --noEmit       # type-check во всех пакетах
pnpm -r lint                    # lint (0 ошибок ожидается)
pnpm format:check               # prettier validation
pnpm -F @eop/dashboard build    # production build конкретного пакета
```

## Tests

Каждый пакет имеет `pnpm test` (Vitest):

| Package | Что покрыто |
| --- | --- |
| `dashboard` | RTL component tests + zod schemas + tz utils |
| `agent` (frontend) | tauri.ts shim mocking `@tauri-apps/api/core.invoke` |
| `browser-extension` | `backend.ts` (config, ingest happy/retry/drop paths) с jsdom + chrome mock |

Rust (`agent/src-tauri`) — `cargo test --lib` (store/preflight). VS Code — `pnpm -F eop-vscode test` (vscode-test, при наличии display).

## Documentation

- Документация **для людей** (деплой, API, архитектура, runbooks) — в [`docs/`](docs/).
- `AGENTS.md` в корне и в пакетах (`backend/AGENTS.md`, `dashboard/AGENTS.md`) — правила для AI, не переносить в `docs/`.
- `README.md` у пакета (`browser-extension/`, `ide-vscode/`) остаётся рядом с кодом.

## Distribution (closed beta)

Артефакты собираются в `.github/workflows/release.yml` на push тэга `v*.*.*` или ручным trigger'ом:

| Артефакт | Платформы | Подпись |
| --- | --- | --- |
| `eop-agent_*.dmg` / `.msi` / `.AppImage` | macOS (universal: aarch64 + x86_64), Windows x64, Linux x64 | Apple Developer ID + notarization, Windows signtool (если есть secrets) — иначе unsigned |
| `eop-vscode.vsix` | VS Code-compatible (Cursor, VS Code, VSCodium) | — |
| `eop-browser-extension.zip` | Chrome / Edge / Brave (MV3) | Unsigned (load unpacked в Developer mode) |

Все артефакты прикрепляются к draft GH Release. Maintainer вручную ревьюит и публикует.

Гайд для beta-тестеров: [`docs/beta-install.md`](docs/beta-install.md).
