/**
 * Категории code provenance ровно в том виде, в каком их пишет attribution
 * worker в ClickHouse-таблицу `attribution_events` (миграция
 * `clickhouse/002_attribution_events.up.sql`). Ключи должны совпадать с
 * Enum8 в схеме — иначе сегмент молча окажется нулевым.
 *
 * Порядок задаёт порядок сегментов доната: сначала написанное человеком,
 * затем неоднозначное, затем AI, затем unknown.
 */
export const PROVENANCE_BUCKETS = [
  { key: "typed", labelKey: "dashboard.provenance_typed", color: "#4ade80" },
  { key: "refactor", labelKey: "dashboard.provenance_refactor", color: "#2dd4bf" },
  { key: "pasted_other", labelKey: "dashboard.provenance_pasted_other", color: "#fbbf24" },
  { key: "ai_inline", labelKey: "dashboard.provenance_ai_inline", color: "hsl(var(--accent))" },
  { key: "pasted_ai", labelKey: "dashboard.provenance_pasted_ai", color: "#60a5fa" },
  { key: "ai_agent", labelKey: "dashboard.provenance_ai_agent", color: "#c084fc" },
  { key: "unknown", labelKey: "dashboard.provenance_unknown", color: "rgba(255,255,255,0.18)" },
] as const;

/**
 * Категории, которые считаются AI-авторством для счётчика в центре доната.
 * Явный набор, а не префикс `ai_`: `pasted_ai` — тоже AI, но под префикс не
 * попадает, а `ai_inline` и `ai_agent` попадают. Прежняя проверка
 * `key.startsWith("ai_")` недосчитывала вставки из AI-чатов.
 */
export const AI_PROVENANCE_KEYS: ReadonlySet<string> = new Set([
  "ai_inline",
  "pasted_ai",
  "ai_agent",
]);
