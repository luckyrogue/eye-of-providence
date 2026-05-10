// global-setup: запускается ОДИН раз перед suite.
//   - resets PG/CH/Redis (clean slate)
//   - ensures backend health (через polling /healthz)
//
// Backend сам прокатывает миграции на startup через AutoMigrate=true.

import { apiHealthz } from "./helpers/api.js";
import { resetAll } from "./helpers/db.js";

async function waitForBackend(): Promise<void> {
  const deadline = Date.now() + 60_000;
  let lastErr: unknown;
  while (Date.now() < deadline) {
    try {
      await apiHealthz();
      return;
    } catch (e) {
      lastErr = e;
      await new Promise((r) => setTimeout(r, 1_000));
    }
  }
  throw new Error(
    `backend not healthy after 60s: ${(lastErr as Error)?.message ?? "unknown"}`,
  );
}

export default async function globalSetup(): Promise<void> {
  // eslint-disable-next-line no-console
  console.log("[e2e] global-setup: waiting for backend...");
  await waitForBackend();

  // eslint-disable-next-line no-console
  console.log("[e2e] global-setup: resetting databases...");
  try {
    resetAll();
  } catch (e) {
    // Если docker exec недоступен (CI runner с пресетным docker compose) —
    // skip. Tests должны быть robust к "грязному" state через unique emails.
    // eslint-disable-next-line no-console
    console.warn(
      `[e2e] global-setup: DB reset skipped (${(e as Error).message})`,
    );
  }

  // eslint-disable-next-line no-console
  console.log("[e2e] global-setup: ready.");
}
