import { test, expect } from "@playwright/test";
import { apiHealthz } from "../helpers/api.js";
test.describe("smoke", () => {
  test("backend /healthz returns ok", async () => {
    const r = await apiHealthz();
    expect(r.status).toMatch(/ok|ready|healthy/i);
  });
  test("dashboard landing loads", async ({ page }) => {
    const resp = await page.goto("/");
    expect(resp?.status()).toBe(200);
    await expect(page.locator("body")).toContainText(/providence|eop|track/i, {
      timeout: 5000,
    });
  });
});
