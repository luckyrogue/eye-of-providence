import { execSync } from "node:child_process";
const PG_DSN =
  process.env.E2E_PG_DSN || "postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable";
const CH_URL = process.env.E2E_CH_URL || "http://localhost:8123";
const CH_USER = process.env.E2E_CH_USER || "eop";
const CH_PASS = process.env.E2E_CH_PASS || "eop_dev";
const CH_DB = process.env.E2E_CH_DB || "eop";
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
  const sql = `TRUNCATE ${PG_TABLES.join(", ")} CASCADE;`;
  execSync(`docker exec eop-postgres psql -U eop -d eop -c "${sql}"`, { stdio: "pipe" });
}
export function resetClickHouse(): void {
  const tables = ["events", "events_hourly_agg", "events_daily_agg"];
  for (const t of tables) {
    try {
      execSync(
        `docker exec eop-clickhouse clickhouse-client --user=${CH_USER} --password=${CH_PASS} --database=${CH_DB} --query="TRUNCATE TABLE IF EXISTS ${t}"`,
        { stdio: "pipe" },
      );
    } catch {}
  }
}
export function resetRedis(): void {
  try {
    execSync(`docker exec eop-redis redis-cli FLUSHALL`, { stdio: "pipe" });
  } catch {}
}
export function resetAll(): void {
  resetPostgres();
  resetClickHouse();
  resetRedis();
}
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
