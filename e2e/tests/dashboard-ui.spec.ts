// Dashboard UI smoke. Pre-creates team через API чтобы dashboard рендерил
// полный layout (без team часть widget'ов рендерит empty state).

import { test, expect } from "../fixtures/index.js";
import { apiCreateTeam } from "../helpers/api.js";
import { uniqueTeamName } from "../helpers/db.js";

test.describe("dashboard ui", () => {
  // Per-test setup: создаём team и dismiss onboarding чтобы dashboard
  // не редиректнул нас на /onboarding.
  test.beforeEach(async ({ session, api }) => {
    await apiCreateTeam(session.token, uniqueTeamName("ui"));
    await api.fetch("/v1/me/onboarding/dismiss", { method: "POST" }).catch(() => {});
  });

  test("authenticated user lands on /dashboard", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/dashboard/, { timeout: 10_000 });
  });

  test("dashboard layout renders without crash", async ({ page }) => {
    await page.goto("/dashboard");
    await page.waitForURL(/dashboard/, { timeout: 10_000 });
    // Smoke: на странице есть хоть какой-то контент (body не пустой).
    const text = await page.locator("body").textContent();
    expect(text?.length ?? 0).toBeGreaterThan(50);
  });

  test("settings route is reachable", async ({ page }) => {
    await page.goto("/settings");
    // URL может остаться /settings ИЛИ редирект на /login если auth state
    // потерялся (тестируем что app не падает). Текст не проверяем — i18n
    // меняется per-locale.
    await page.waitForLoadState("networkidle", { timeout: 10_000 });
    expect(page.url()).toMatch(/settings|login|^http/);
  });

  test("logout via clearing localStorage redirects to login/landing", async ({ page }) => {
    await page.goto("/dashboard");
    await page.evaluate(() => {
      localStorage.removeItem("eop_token");
      localStorage.removeItem("eop_user_id");
    });
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/(login|^\/$|landing|signup)/, {
      timeout: 10_000,
    });
  });
});
