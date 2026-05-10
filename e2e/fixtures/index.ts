// Custom Playwright fixtures: extend base test с авто-инжектированной
// session'ой (для большинства тестов, где login flow не интересен) и
// API-клиентом с уже-готовым токеном.
//
// Pattern (см. https://playwright.dev/docs/test-fixtures):
//
//   import { test, expect } from "@e2e/fixtures";
//   test("...", async ({ page, session, api }) => { ... });

import { test as baseTest, expect } from "@playwright/test";
import {
  createApiClient,
  type ApiClient,
} from "../helpers/api.js";
import { createSession, type Session } from "../helpers/auth.js";

type Fixtures = {
  session: Session;
  api: ApiClient;
};

export const test = baseTest.extend<Fixtures>({
  session: async ({ context }, use) => {
    const session = await createSession(context);
    await use(session);
  },
  api: async ({ session }, use) => {
    await use(createApiClient(session.token));
  },
});

export { expect };
