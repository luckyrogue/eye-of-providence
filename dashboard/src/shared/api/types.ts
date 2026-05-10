// High-level type aliases поверх openapi.d.ts.
//
// Generated файл (openapi.d.ts) описывает каждый endpoint через nested
// paths-структуру. Прямое использование `paths["/v1/me"]["get"]["responses"]
// ["200"]["content"]["application/json"]` боль; этот модуль нормализует
// в плоские типы которые удобно импортировать в entities/.
//
// Регенерация: `pnpm openapi:gen` после изменений docs/api/openapi.yaml.

import type { components } from "./openapi";

export type Schemas = components["schemas"];

// Удобные re-export'ы. Имена совпадают с YAML-определениями.
export type AuthResponse = Schemas["AuthResponse"];
export type Me = Schemas["Me"];
export type Insight = Schemas["Insight"];
export type APIToken = Schemas["APIToken"];
export type Webhook = Schemas["Webhook"];
export type Event = Schemas["Event"];
export type LangCell = Schemas["LangCell"];
export type TrendPoint = Schemas["TrendPoint"];
