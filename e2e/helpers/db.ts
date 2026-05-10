// DB cleanup helper для tests. Подключается к локальному PG/CH/Redis из
// docker compose (infra/docker-compose.yml).
//
// Используется в global-setup для clean slate ДО прогона suite. Не для
// per-test cleanup — это slow (3 DB connect'a). Tests должны быть
// идемпотентны через unique emails / team names.
//
// Если нужно reset state в середине suite — вызывать `resetAll()` из
// отдельного `beforeAll` в spec файле где это требуется.

import { execSync } from "node:child_process";

const PG_DSN =
  process.env.E2E_PG_DSN || "postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable";
const CH_URL = process.env.E2E_CH_URL || "http://localhost:8123";
const CH_USER = process.env.E2E_CH_USER || "eop";
const CH_PASS = process.env.E2E_CH_PASS || "eop_dev";
const CH_DB = process.env.E2E_CH_DB || "eop";

// Postgres tables в порядке зависимостей (parent → child cascade — но мы явно
// чистим всё для idempotency).
const PG_TABLES = [
  "team_payments",
  "password_resets",
  "team_invites",
  "commits",
  "projects",
  "team_members",
  "teams",
  "api_tokens",
  "push_subscriptions",
  "webhook_subscriptions",
  "users",
];

export function resetPostgres(): void {
  // Через psql напрямую — pg client'а в e2e package нет (не хотим тащить
  // ещё одну зависимость). docker exec → psql внутри контейнера.
  const sql = `TRUNCATE ${PG_TABLES.join(", ")} CASCADE;`;
  execSync(`docker exec eop-postgres psql -U eop -d eop -c "${sql}"`, { stdio: "pipe" });
}

export function resetClickHouse(): void {
  // ClickHouse events table — DROP не нужен, TRUNCATE достаточно. Aggregates
  // (events_hourly_agg, events_daily_agg) — тоже clean чтобы dashboard не
  // подхватил остаточные данные между тестами.
  const tables = ["events", "events_hourly_agg", "events_daily_agg"];
  for (const t of tables) {
    try {
      execSync(
        `docker exec eop-clickhouse clickhouse-client --user=${CH_USER} --password=${CH_PASS} --database=${CH_DB} --query="TRUNCATE TABLE IF EXISTS ${t}"`,
        { stdio: "pipe" },
      );
    } catch {
      // Table может не существовать на первом запуске (миграции ещё не прошли).
      // Backend сам прокатит миграции через AutoMigrate на startup.
    }
  }
}

export function resetRedis(): void {
  // Кешированные responses в `eop:cache:*` — очистить.
  try {
    execSync(`docker exec eop-redis redis-cli FLUSHALL`, { stdio: "pipe" });
  } catch {
    // Redis optional — backend gracefully degrades если недоступен.
  }
}

export function resetAll(): void {
  resetPostgres();
  resetClickHouse();
  resetRedis();
}

// uniqueEmail — каждый test генерит свой email чтобы между прогонами не
// сталкивались. Шаблон: `e2e-{timestamp}-{rand}@local.test`. Локальная зона
// `.test` reserved (RFC 2606) — не разойдётся в DNS даже случайно.
export function uniqueEmail(prefix = "user"): string {
  const ts = Date.now();
  const rand = Math.random().toString(36).slice(2, 8);
  return `e2e-${prefix}-${ts}-${rand}@local.test`;
}

export function uniqueTeamName(prefix = "team"): string {
  const ts = Date.now();
  const rand = Math.random().toString(36).slice(2, 6);
  return `${prefix}-${ts}-${rand}`;
}
