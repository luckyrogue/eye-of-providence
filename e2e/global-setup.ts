import { apiHealthz } from "./helpers/api.js";
import { resetAll } from "./helpers/db.js";
async function waitForBackend(): Promise<void> {
  const deadline = Date.now() + 60000;
  let lastErr: unknown;
  while (Date.now() < deadline) {
    try {
      await apiHealthz();
      return;
    } catch (e) {
      lastErr = e;
      await new Promise((r) => setTimeout(r, 1000));
    }
  }
  throw new Error(`backend not healthy after 60s: ${(lastErr as Error)?.message ?? "unknown"}`);
}
export default async function globalSetup(): Promise<void> {
  console.log("[e2e] global-setup: waiting for backend...");
  await waitForBackend();
  console.log("[e2e] global-setup: resetting databases...");
  try {
    resetAll();
  } catch (e) {
    console.warn(`[e2e] global-setup: DB reset skipped (${(e as Error).message})`);
  }
  console.log("[e2e] global-setup: ready.");
}
