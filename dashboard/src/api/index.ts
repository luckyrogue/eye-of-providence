// Объединённый barrel: реэкспорт всех модулей api.
// Существующие импорты `from "./api"` продолжают работать благодаря
// TypeScript module-resolution (api.ts заменился на api/index.ts).

export { AUTH_FAILED_EVENT } from "../lib/http";

export * from "./auth";
export * from "./me";
export * from "./teams";
export * from "./events";
export * from "./reports";
export * from "./admin";
