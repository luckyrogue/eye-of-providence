// Webhooks management (CRUD). Outbound delivery не тестируем (требует
// внешнего receiver) — это есть в unit-тестах backend/internal/webhooks.

import { test, expect } from "../fixtures/index.js";
import { ApiError } from "../helpers/api.js";
import type { Webhook } from "../helpers/types.js";

interface CreateResp {
  secret: string;
  webhook: Webhook;
}

test.describe("webhooks", () => {
  test("create → list → delete", async ({ api }) => {
    // Create
    const r = await api.fetch<CreateResp>("/v1/me/webhooks", {
      method: "POST",
      body: JSON.stringify({
        url: "https://webhook.test.local/receive",
        events: ["commit.ingested"],
        format: "json",
      }),
    });
    expect(r.secret).toMatch(/^[a-f0-9]{32,}$/);
    expect(r.webhook.id).toMatch(/^[0-9a-f-]{36}$/);

    // List
    const list = await api.fetch<{ webhooks: Webhook[] }>("/v1/me/webhooks");
    expect(list.webhooks.find((w) => w.id === r.webhook.id)).toBeDefined();

    // Delete
    await api.fetch(`/v1/me/webhooks/${r.webhook.id}`, { method: "DELETE" });
    const after = await api.fetch<{ webhooks: Webhook[] }>("/v1/me/webhooks");
    expect(after.webhooks.find((w) => w.id === r.webhook.id)).toBeUndefined();
  });

  test("delete non-existent webhook → webhook_not_found", async ({ api }) => {
    try {
      await api.fetch("/v1/me/webhooks/00000000-0000-0000-0000-000000000000", {
        method: "DELETE",
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(404);
      expect(err.code).toBe("webhook_not_found");
    }
  });

  test("invalid webhook URL → create_failed", async ({ api }) => {
    try {
      await api.fetch("/v1/me/webhooks", {
        method: "POST",
        body: JSON.stringify({
          url: "not-a-url",
          events: ["commit.ingested"],
          format: "json",
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("create_failed");
    }
  });
});
