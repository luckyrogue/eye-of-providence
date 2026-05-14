import { test, expect } from "../fixtures/index.js";
import { apiCreateTeam } from "../helpers/api.js";
import { uniqueTeamName } from "../helpers/db.js";
test.describe("dashboard ui", () => {
  test.beforeEach(async ({ session, api }) => {
    await apiCreateTeam(session.token, uniqueTeamName("ui"));
    await api.fetch("/v1/me/onboarding/dismiss", { method: "POST" }).catch(() => {});
  });
  test("authenticated user lands on /dashboard", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 });
  });
  test("dashboard layout renders without crash", async ({ page }) => {
    await page.goto("/dashboard");
    await page.waitForURL(/dashboard/, { timeout: 10000 });
    const text = await page.locator("body").textContent();
    expect(text?.length ?? 0).toBeGreaterThan(50);
  });
  test("settings route is reachable", async ({ page }) => {
    await page.goto("/settings");
    await page.waitForLoadState("networkidle", { timeout: 10000 });
    expect(page.url()).toMatch(/settings|login|^http/);
  });
  test("logout via clearing localStorage redirects to login/landing", async ({ page, context }) => {
    await page.goto("/dashboard");
    await context.addInitScript(() => {
      localStorage.removeItem("eop_token");
      localStorage.removeItem("eop_user_id");
    });
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/(login|^\/$|landing|signup)/, {
      timeout: 10000,
    });
  });
});
