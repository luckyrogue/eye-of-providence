// Dashboard UI smoke: после login юзер видит dashboard route, в navigation
// присутствуют ключевые ссылки. Проверка регрессий FSD-migration (router /
// auth flow).

import { test, expect } from "../fixtures/index.js";

test.describe("dashboard ui", () => {
  test("authenticated user lands on /dashboard or /onboarding", async ({
    page,
    session: _session,
  }) => {
    await page.goto("/dashboard");
    // Если onboarding не пройден, может редиректнуть; принимаем оба.
    await expect(page).toHaveURL(/(dashboard|onboarding)/, {
      timeout: 10_000,
    });
  });

  test("dashboard has nav links to team / settings", async ({ page }) => {
    await page.goto("/dashboard");
    // Принимаем редирект на /onboarding если бэк так решил.
    await page.waitForURL(/(dashboard|onboarding)/);
    if (!page.url().includes("/dashboard")) {
      // Skip — мы на onboarding'е, что тоже валидно. Test покрывает
      // post-onboarding state.
      test.skip(true, "redirected to onboarding (no nav yet)");
      return;
    }
    // Nav может быть burger menu на мобилке — проверяем что хоть один
    // navigation link присутствует в DOM.
    const links = await page.locator("a[href*='/team'], a[href*='/settings']").count();
    expect(links).toBeGreaterThan(0);
  });

  test("settings page loads", async ({ page }) => {
    await page.goto("/settings");
    await expect(page).toHaveURL(/settings/);
    // На странице должен быть какой-то заголовок (Settings / Настройки).
    await expect(page.locator("body")).toContainText(/settings|настройки/i);
  });

  test("logout via clearing localStorage redirects to login", async ({
    page,
  }) => {
    await page.goto("/dashboard");
    await page.evaluate(() => {
      localStorage.removeItem("eop_token");
      localStorage.removeItem("eop_user_id");
    });
    await page.goto("/dashboard");
    // Без token — должно перекинуть на /login или /.
    await expect(page).toHaveURL(/(login|^\/$|landing)/, { timeout: 10_000 });
  });
});
