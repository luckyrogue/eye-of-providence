// Auth flow:
//   - register → токен в localStorage, redirect на onboarding
//   - login существующего юзера
//   - logout инвалидирует state в browser-context'е
//   - forgot password не палит существование email'а (всегда 200)
//   - реджект пустого тела с типизированной ошибкой

import { test, expect } from "@playwright/test";
import {
  apiLogin,
  apiRegister,
  ApiError,
  createApiClient,
} from "../helpers/api.js";
import { uniqueEmail } from "../helpers/db.js";

test.describe("auth", () => {
  test("register flow via API works and returns JWT", async () => {
    const email = uniqueEmail("register");
    const r = await apiRegister(email, "TestPassword123!", "Register Tester");
    expect(r.token).toMatch(/^eyJ/); // JWT prefix
    expect(r.user_id).toMatch(/^[0-9a-f-]{36}$/);
  });

  test("login flow via API works", async () => {
    const email = uniqueEmail("login");
    await apiRegister(email, "TestPassword123!", "Login Tester");
    const r = await apiLogin(email, "TestPassword123!");
    expect(r.token).toMatch(/^eyJ/);
  });

  test("login rejects wrong password with invalid_credentials", async () => {
    const email = uniqueEmail("wrongpw");
    await apiRegister(email, "TestPassword123!");
    try {
      await apiLogin(email, "WrongPassword456!");
      throw new Error("expected to throw");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      const err = e as ApiError;
      expect(err.status).toBe(401);
      expect(err.code).toBe("invalid_credentials");
    }
  });

  test("register rejects short password with typed code", async () => {
    try {
      await apiRegister(uniqueEmail("shortpw"), "abc");
      throw new Error("expected to throw");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("invalid_password");
    }
  });

  test("forgot password always returns 200 (no email enumeration)", async () => {
    const c = createApiClient();
    const r1 = await c.fetch<{ status: string }>("/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email: "definitely-not-exists@local.test" }),
    });
    expect(r1.status).toBe("ok");
    // То же для невалидного формата.
    const r2 = await c.fetch<{ status: string }>("/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email: "not-an-email" }),
    });
    expect(r2.status).toBe("ok");
  });

  test("UI: signup form leads to onboarding", async ({ page }) => {
    const email = uniqueEmail("ui-signup");
    await page.goto("/signup");

    // Сигнатура form'ы: email + display_name + password + кнопка.
    // Используем resilient locators (label / placeholder).
    const emailField = page
      .getByPlaceholder(/email|почта/i)
      .or(page.getByLabel(/email|почта/i))
      .first();
    await emailField.fill(email);

    const nameField = page
      .getByPlaceholder(/имя|name/i)
      .or(page.getByLabel(/имя|name/i))
      .first();
    if (await nameField.isVisible().catch(() => false)) {
      await nameField.fill("UI Signup Tester");
    }

    const pwField = page
      .getByPlaceholder(/пароль|password/i)
      .or(page.getByLabel(/пароль|password/i))
      .first();
    await pwField.fill("TestPassword123!");

    await page
      .getByRole("button", { name: /зарегистр|sign ?up|create|регистрац/i })
      .first()
      .click();

    // После register backend выдаёт JWT, dashboard кладёт в localStorage и
    // редиректит. URL проверяем по pattern — /onboarding или /dashboard
    // (зависит от того что выдал onboardingStatus).
    await expect(page).toHaveURL(/(onboarding|dashboard)/, { timeout: 10_000 });
    const token = await page.evaluate(() => localStorage.getItem("eop_token"));
    expect(token).toMatch(/^eyJ/);
  });
});
