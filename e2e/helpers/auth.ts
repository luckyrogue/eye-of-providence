// Auth helpers для browser context'а: записать JWT в localStorage чтобы
// dashboard сразу узнал юзера, либо выполнить полный UI register flow.
//
// Browser-side: dashboard читает токен через `localStorage.getItem("eop_token")`
// (см. dashboard/src/entities/session). Мы записываем туда тот же ключ.

import type { BrowserContext, Page } from "@playwright/test";
import { apiRegister, apiLogin, type AuthResponse } from "./api.js";
import { uniqueEmail } from "./db.js";

const TOKEN_KEY = "eop_token";
const USER_ID_KEY = "eop_user_id";

// authenticatedSession — создаёт юзера через API, кладёт токен в localStorage,
// возвращает credentials. Используется в большинстве тестов, где UI-register
// flow не предмет тестирования.
export interface Session extends AuthResponse {
  email: string;
  password: string;
}

export async function createSession(
  context: BrowserContext,
  opts: { prefix?: string; password?: string } = {},
): Promise<Session> {
  const email = uniqueEmail(opts.prefix);
  const password = opts.password ?? "TestPassword123!";
  const resp = await apiRegister(email, password, email.split("@")[0]);

  await context.addInitScript(
    (args: { token: string; userID: string; tokenKey: string; userKey: string }) => {
      localStorage.setItem(args.tokenKey, args.token);
      localStorage.setItem(args.userKey, args.userID);
    },
    {
      token: resp.token,
      userID: resp.user_id,
      tokenKey: TOKEN_KEY,
      userKey: USER_ID_KEY,
    },
  );

  return { ...resp, email, password };
}

// loginExistingViaAPI — login через API + inject в context. Если у тебя уже
// есть credentials (например first-user bootstrap), это быстрее чем UI.
export async function loginExistingViaAPI(
  context: BrowserContext,
  email: string,
  password: string,
): Promise<Session> {
  const resp = await apiLogin(email, password);
  await context.addInitScript(
    (args: { token: string; userID: string; tokenKey: string; userKey: string }) => {
      localStorage.setItem(args.tokenKey, args.token);
      localStorage.setItem(args.userKey, args.userID);
    },
    {
      token: resp.token,
      userID: resp.user_id,
      tokenKey: TOKEN_KEY,
      userKey: USER_ID_KEY,
    },
  );
  return { ...resp, email, password };
}

// uiRegister — полный UI flow через signup form. Для тестов самой регистрации.
export async function uiRegister(
  page: Page,
  email: string,
  password: string,
  displayName?: string,
): Promise<void> {
  await page.goto("/signup");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/имя|display name/i).fill(displayName ?? email.split("@")[0]);
  await page.getByLabel(/пароль|password/i).fill(password);
  await page
    .getByRole("button", { name: /зарегистрироваться|register|sign up/i })
    .click();
}

export async function uiLogin(
  page: Page,
  email: string,
  password: string,
): Promise<void> {
  await page.goto("/login");
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/пароль|password/i).fill(password);
  await page
    .getByRole("button", { name: /войти|login|sign in/i })
    .click();
}

export async function uiLogout(page: Page): Promise<void> {
  await page.evaluate((key: string) => {
    localStorage.removeItem(key);
  }, TOKEN_KEY);
  await page.goto("/login");
}
