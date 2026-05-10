# @eop/ui — UI primitives library

См. корневой `AGENTS.md` для общих frontend-правил.

## Особенности

- `widgets/*` — доменные композиции (StatTile, PlanBadge, EmptyState, PageHeader, Stepper, DangerZone).
- `features/*` — императивные сценарии с собственным контекстом (`useConfirm`, `PromptDialog`).
- `shared/lib/*` — утилиты (`cn`).
- `shared/styles.css` — общие CSS-переменные shadcn-темы.
- Никаких dependencies на `dashboard`/`agent`/`browser-extension`.
- i18n-агностично — caller передаёт строки через props (например, `copyLabel` у `SecretField`).
- Public API — только через `src/index.ts`. Внутренние пути не импортируем из потребителей.

## Структура `shared/ui`

- **`shared/ui/input/`** — `Input`, `InputField` (RHF + `Form`), типы.
- **`shared/ui/select/`** — примитивы Radix Select, `SimpleSelect`, `SelectField` (RHF), типы.
- **`shared/ui/button/`** — `Button`, `IconButton`.
- **`shared/ui/avatar/`** — `Avatar`, утилита `getInitials`.
- Отдельные файлы в корне `shared/ui/` (например `checkbox.tsx`, `form.tsx`) — по мере необходимости; новые примитивы по возможности кластеризовать в папки как выше.

## Локальный ESLint

`@eop/ui` сам — авторы `Button`/`IconButton`/`SecretField`/`Input`/`Select` и т.д., поэтому `no-restricted-syntax` для нас выключен в `eslint.config.js`. В `eslint.base.js` для потребителей запрещены голые `<input>`/`<select>` (см. корневой `AGENTS.md`).
