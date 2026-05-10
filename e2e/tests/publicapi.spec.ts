// Public API (read-only, /v1/public/*). Авторизация через JWT с
// RequireScope("read", "admin"). Endpoints стабильны для интеграций.

import { test, expect } from "../fixtures/index.js";
import { createApiClient, ApiError } from "../helpers/api.js";

test.describe("public api", () => {
  test("events endpoint returns events + count + limit", async ({ api }) => {
    const r = await api.fetch<{
      events: unknown[];
      count: number;
      limit: number;
    }>("/v1/public/events?limit=10");
    expect(Array.isArray(r.events)).toBe(true);
    expect(r.limit).toBe(10);
  });

  test("summary endpoint returns categories", async ({ api }) => {
    const r = await api.fetch<{ categories: unknown; days: number }>(
      "/v1/public/summary?days=7",
    );
    expect(r.days).toBe(7);
    expect(r.categories).toBeDefined();
  });

  test("languages endpoint returns cells", async ({ api }) => {
    const r = await api.fetch<{ cells: unknown[]; days: number }>(
      "/v1/public/languages?days=30",
    );
    expect(Array.isArray(r.cells)).toBe(true);
  });

  test("trend endpoint returns points", async ({ api }) => {
    const r = await api.fetch<{ points: unknown[] }>(
      "/v1/public/trend?days=7&tz=UTC",
    );
    expect(Array.isArray(r.points)).toBe(true);
  });

  test("read-only token has access", async ({ api }) => {
    const token = await api.fetch<{ token: string }>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-public",
        scope: "read",
        ttl_days: 1,
      }),
    });
    const readC = createApiClient(token.token);
    const r = await readC.fetch<{ events: unknown[] }>(
      "/v1/public/events?limit=5",
    );
    expect(Array.isArray(r.events)).toBe(true);
  });

  test("write:ingest scope cannot access read endpoints", async ({ api }) => {
    const token = await api.fetch<{ token: string }>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-ingest-only",
        scope: "write:ingest",
        ttl_days: 1,
      }),
    });
    const ingestOnly = createApiClient(token.token);
    try {
      await ingestOnly.fetch("/v1/public/events?limit=5");
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("scope_insufficient");
    }
  });
});
