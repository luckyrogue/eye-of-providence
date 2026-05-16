// DB cleanup helper для tests. Подключается к локальному PG/CH/Redis из
// docker compose (infra/docker-compose.yml).
//
// Используется в global-setup для clean slate ДО прогона suite. Не для
// per-test cleanup — это slow (3 DB connect'a). Tests должны быть
// идемпотентны через unique emails / team names.
//
// Если нужно reset state в середине suite — вызывать `resetAll()` из
// отдельного `beforeAll` в spec файле где это требуется.

import { execFileSync } from "node:child_process";

// PG/CH connect-strings раньше пробрасывались внутрь shell-команды через
// `execSync(string)`. CodeQL js/indirect-command-line-injection (#137) — если
// E2E_PG_DSN/E2E_CH_URL пришёл из process.env с метасимволами, он встроится
// в шелл-команду как есть. Переехали на execFileSync с array-args (без shell):
// connect-info теперь не нужна — psql/clickhouse-client внутри контейнеров уже
// сконфигурированы под локальный compose-stack. Если в будущем понадобится
// разный host/port, передаём через --host/--port отдельными элементами массива.
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
  "sso_states",
  "sso_configs",
  "teams",
  "api_tokens",
  "push_subscriptions",
  "webhooks",
  "users",
];

export function resetPostgres(): void {
  // Через psql напрямую — pg client'а в e2e package нет (не хотим тащить
  // ещё одну зависимость). docker exec → psql внутри контейнера.
  // execFileSync(array) — без shell, никакая инъекция через PG_TABLES не
  // пройдёт (имена таблиц hardcoded, но array-form всё равно безопаснее).
  const sql = `TRUNCATE ${PG_TABLES.join(", ")} CASCADE;`;
  execFileSync(
    "docker",
    ["exec", "eop-dev-postgres", "psql", "-U", "eop", "-d", "eop", "-c", sql],
    { stdio: "pipe" },
  );
}

export function resetClickHouse(): void {
  // ClickHouse events table — DROP не нужен, TRUNCATE достаточно. Aggregates
  // (events_hourly_agg, events_daily_agg) — тоже clean чтобы dashboard не
  // подхватил остаточные данные между тестами.
  const tables = ["events", "events_hourly_agg", "events_daily_agg"];
  for (const t of tables) {
    try {
      execFileSync(
        "docker",
        [
          "exec",
          "eop-dev-clickhouse",
          "clickhouse-client",
          `--user=${CH_USER}`,
          `--password=${CH_PASS}`,
          `--database=${CH_DB}`,
          `--query=TRUNCATE TABLE IF EXISTS ${t}`,
        ],
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
    execFileSync("docker", ["exec", "eop-dev-redis", "redis-cli", "FLUSHALL"], { stdio: "pipe" });
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
