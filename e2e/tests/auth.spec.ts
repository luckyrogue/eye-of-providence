import { test, expect } from "@playwright/test";
import { apiLogin, apiRegister, ApiError, createApiClient } from "../helpers/api.js";
import { uniqueEmail } from "../helpers/db.js";
test.describe("auth", () => {
  test("register flow via API works and returns JWT", async () => {
    const email = uniqueEmail("register");
    const r = await apiRegister(email, "TestPassword123!", "Register Tester");
    expect(r.token).toMatch(/^eyJ/);
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
    const r1 = await c.fetch<{
      status: string;
    }>("/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email: "definitely-not-exists@local.test" }),
    });
    expect(r1.status).toBe("ok");
    const r2 = await c.fetch<{
      status: string;
    }>("/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email: "not-an-email" }),
    });
    expect(r2.status).toBe("ok");
  });
  test.skip("UI: signup form leads to onboarding", async () => {});
});
