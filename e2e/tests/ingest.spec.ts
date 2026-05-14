import { test, expect } from "../fixtures/index.js";
import type { Event } from "../helpers/types.js";
import { ApiError } from "../helpers/api.js";
function validEvent(overrides: Partial<Event> = {}): Event {
  return {
    ts: new Date().toISOString(),
    app_bundle: "com.test.editor",
    source: "ide",
    category: "ai",
    duration_ms: 5000,
    chars_in: 100,
    lines_added: 5,
    lines_removed: 0,
    file_lang: "typescript",
    ...overrides,
  };
}
test.describe("ingest", () => {
  test("happy path: 3 events accepted", async ({ api }) => {
    const r = await api.fetch<{
      accepted: number;
      rejected: number;
    }>("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [validEvent(), validEvent(), validEvent()],
      }),
    });
    expect(r.accepted).toBe(3);
    expect(r.rejected).toBe(0);
  });
  test("invalid event (bad category) rejected, valid ones accepted", async ({ api }) => {
    const r = await api.fetch<{
      accepted: number;
      rejected: number;
    }>("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [validEvent(), validEvent({ category: "bogus_category" }), validEvent()],
      }),
    });
    expect(r.accepted).toBe(2);
    expect(r.rejected).toBe(1);
  });
  test("batch over 5000 events → 413 batch_too_large", async ({ api }) => {
    const events = Array.from({ length: 5001 }, () => validEvent());
    try {
      await api.fetch("/v1/ingest", {
        method: "POST",
        body: JSON.stringify({ events }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(413);
      expect(err.code).toBe("batch_too_large");
      expect(
        (
          err.body as {
            max_batch?: number;
          }
        )?.max_batch,
      ).toBe(5000);
    }
  });
  test("ingested event appears in /v1/events/recent", async ({ api }) => {
    const marker = `e2e-marker-${Date.now()}`;
    await api.fetch("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [validEvent({ app_bundle: marker })],
      }),
    });
    let found = false;
    for (let attempt = 0; attempt < 10; attempt++) {
      const out = await api.fetch<{
        events: Event[];
      }>("/v1/events/recent?limit=50");
      if (out.events.some((e) => e.app_bundle === marker)) {
        found = true;
        break;
      }
      await new Promise((r) => setTimeout(r, 500));
    }
    expect(found).toBe(true);
  });
});
