import { defineConfig, devices } from "@playwright/test";

// Ports:
//   - backend (Go cmd/api):  http://localhost:8080
//   - dashboard (vite preview): http://localhost:4173
//
// CI prerequisites (handled by .github/workflows/e2e.yml):
//   - PG/CH/Redis up (docker compose -f infra/docker-compose.yml up -d)
//   - backend env vars (EOP_*, see global-setup.ts comment)
//   - `pnpm -F @eop/dashboard build` с VITE_BACKEND_URL=http://localhost:8080
//
// Local: `pnpm e2e:up && pnpm e2e:test` (см. root package.json scripts).

const PORT_DASHBOARD = 4173;
const PORT_BACKEND = 8080;
const CI = !!process.env.CI;

export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  expect: { timeout: 5_000 },

  globalSetup: "./global-setup.ts",
  fullyParallel: false, // shared DB state → serial для предсказуемости
  forbidOnly: CI,
  retries: CI ? 1 : 0,
  workers: 1, // single worker, чтобы tests с DB seeding не stomp'или друг друга

  reporter: CI
    ? [["github"], ["html", { open: "never" }], ["list"]]
    : [["html", { open: "on-failure" }], ["list"]],

  use: {
    baseURL: `http://localhost:${PORT_DASHBOARD}`,
    trace: CI ? "retain-on-failure" : "on-first-retry",
    screenshot: "only-on-failure",
    video: CI ? "retain-on-failure" : "off",
    actionTimeout: 10_000,
    navigationTimeout: 15_000,
    extraHTTPHeaders: {
      "X-E2E-Test": "1",
    },
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    // Firefox/WebKit можно расширить когда suite стабилизируется.
  ],

  // Запускаем backend + dashboard через playwright webServer чтобы локально
  // не нужно было руками поднимать процессы. В CI workflow поднимает их
  // отдельно и эти blocks пропустятся через `reuseExistingServer: true`.
  webServer: [
    {
      command:
        "cd ../backend && EOP_HTTP_ADDR=:8080 EOP_ENV=development EOP_INVITE_ONLY=false EOP_ENABLE_DEV_TOKEN=true EOP_ALLOWED_ORIGINS=http://localhost:4173 EOP_AUTO_MIGRATE=true EOP_JWT_SECRET=e2e_test_secret_at_least_32_chars_xx EOP_POSTGRES_DSN=postgres://eop:eop_dev@localhost:5432/eop?sslmode=disable EOP_CLICKHOUSE_DSN=clickhouse://eop:eop_dev@localhost:9000/eop EOP_REDIS_ADDR=localhost:6379 EOP_REPORTS_CRON_SEC=0 go run ./cmd/api",
      url: `http://localhost:${PORT_BACKEND}/healthz`,
      timeout: 120_000,
      reuseExistingServer: !CI,
      stdout: "pipe",
      stderr: "pipe",
    },
    {
      command:
        "cd ../dashboard && VITE_BACKEND_URL=http://localhost:8080 pnpm preview --host 127.0.0.1 --port 4173 --strictPort",
      url: `http://localhost:${PORT_DASHBOARD}`,
      timeout: 60_000,
      reuseExistingServer: !CI,
      stdout: "pipe",
      stderr: "pipe",
    },
  ],
});
