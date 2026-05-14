import { test, expect } from "../fixtures/index.js";
import type { Event } from "../helpers/types.js";
test.describe("insights", () => {
  test("returns insights array after ingest", async ({ api }) => {
    const events: Event[] = Array.from({ length: 30 }, (_, i) => ({
      ts: new Date(Date.now() - i * 60000).toISOString(),
      app_bundle: "com.test.insights",
      source: "ide",
      category: i % 2 ? "ai" : "manual",
      file_lang: "typescript",
      duration_ms: 60000,
      chars_in: 100,
      lines_added: 3,
      lines_removed: 1,
    }));
    await api.fetch("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({ events }),
    });
    let insights: unknown[] = [];
    for (let attempt = 0; attempt < 10; attempt++) {
      const r = await api.fetch<{
        insights: unknown[];
      }>("/v1/me/insights");
      if (r.insights) {
        insights = r.insights;
        break;
      }
      await new Promise((r) => setTimeout(r, 500));
    }
    expect(Array.isArray(insights)).toBe(true);
  });
});
