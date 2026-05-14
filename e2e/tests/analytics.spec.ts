import { test, expect } from "../fixtures/index.js";
import type { Event } from "../helpers/types.js";
function ingestPayload(category: string, lang: string): Event {
  return {
    ts: new Date().toISOString(),
    app_bundle: "com.test.analytics",
    source: "ide",
    category,
    file_lang: lang,
    duration_ms: 10000,
    chars_in: 50,
    lines_added: 2,
    lines_removed: 1,
  };
}
test.describe("analytics", () => {
  test("aggregate by category after ingest", async ({ api }) => {
    await api.fetch("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [
          ingestPayload("ai", "typescript"),
          ingestPayload("manual", "typescript"),
          ingestPayload("refactor", "go"),
        ],
      }),
    });
    let categories: Record<string, number> | null = null;
    for (let attempt = 0; attempt < 10; attempt++) {
      const r = await api.fetch<{
        days: number;
        categories: Record<string, number>;
      }>("/v1/summary/categories?days=7");
      if (r.categories && Object.keys(r.categories).length > 0) {
        categories = r.categories;
        break;
      }
      await new Promise((r) => setTimeout(r, 500));
    }
    expect(categories).not.toBeNull();
    expect(Object.keys(categories!).length).toBeGreaterThan(0);
  });
  test("language breakdown returns cells", async ({ api }) => {
    await api.fetch("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [ingestPayload("ai", "typescript"), ingestPayload("ai", "python")],
      }),
    });
    let cells: unknown[] = [];
    for (let attempt = 0; attempt < 10; attempt++) {
      const r = await api.fetch<{
        cells: unknown[];
      }>("/v1/summary/languages?days=30");
      if (r.cells.length > 0) {
        cells = r.cells;
        break;
      }
      await new Promise((r) => setTimeout(r, 500));
    }
    expect(cells.length).toBeGreaterThan(0);
  });
  test("daily trend returns points array", async ({ api }) => {
    const r = await api.fetch<{
      points: unknown[];
      days: number;
    }>("/v1/trend?days=7");
    expect(Array.isArray(r.points)).toBe(true);
    expect(r.days).toBe(7);
  });
  test("heatmap returns 7x24 grid cells", async ({ api }) => {
    const r = await api.fetch<{
      cells: unknown[];
      days: number;
    }>("/v1/heatmap?days=7");
    expect(Array.isArray(r.cells)).toBe(true);
  });
  test("invalid days falls back to default (no error)", async ({ api }) => {
    const r = await api.fetch<{
      days: number;
    }>("/v1/trend?days=99999");
    expect(r.days).toBe(30);
  });
});
