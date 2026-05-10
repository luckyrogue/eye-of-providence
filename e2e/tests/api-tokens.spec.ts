// API tokens lifecycle:
//   - create token с scope=write:ingest → plaintext отдан 1 раз
//   - use токен для ingest → events accepted
//   - read-only scope не может ingest'ить (scope_insufficient)
//   - delete токен → больше не работает
//   - API token не может создавать другие токены (jwt_required guard)

import { test, expect } from "../fixtures/index.js";
import { createApiClient, ApiError } from "../helpers/api.js";
import type { APIToken } from "../helpers/types.js";

interface CreateTokenResp {
  token: string;
  metadata: APIToken;
}

test.describe("api-tokens", () => {
  test("create token with write:ingest scope, use it to ingest", async ({ api }) => {
    const r = await api.fetch<CreateTokenResp>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-write-token",
        scope: "write:ingest",
        ttl_days: 7,
      }),
    });
    expect(r.token).toMatch(/^eop_[a-zA-Z0-9_-]+/);
    expect(r.metadata.scope).toBe("write:ingest");
    expect(r.metadata.prefix).toMatch(/^eop_/);

    // Используем token для ingest.
    const ingestC = createApiClient(r.token);
    const ingestResp = await ingestC.fetch<{ accepted: number }>("/v1/ingest", {
      method: "POST",
      body: JSON.stringify({
        events: [
          {
            app_bundle: "com.test.bundle",
            source: "ide",
            category: "manual",
            duration_ms: 1000,
          },
        ],
      }),
    });
    expect(ingestResp.accepted).toBe(1);
  });

  test("read-only token cannot ingest (scope_insufficient)", async ({ api }) => {
    const r = await api.fetch<CreateTokenResp>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-read-only",
        scope: "read",
        ttl_days: 1,
      }),
    });
    const readC = createApiClient(r.token);
    try {
      await readC.fetch("/v1/ingest", {
        method: "POST",
        body: JSON.stringify({
          events: [
            {
              app_bundle: "x",
              source: "ide",
              category: "manual",
              duration_ms: 1,
            },
          ],
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("scope_insufficient");
      // RFC 7807 extension fields на верхнем уровне response.
      const body = err.body as { actual_scope?: string; required_scope?: string[] };
      expect(body.actual_scope).toBe("read");
      expect(body.required_scope).toContain("write:ingest");
    }
  });

  test("API token cannot create another token (jwt_required guard)", async ({ api }) => {
    const r = await api.fetch<CreateTokenResp>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-bootstrap-attempt",
        scope: "admin",
        ttl_days: 1,
      }),
    });
    const adminC = createApiClient(r.token);
    try {
      await adminC.fetch("/v1/me/tokens", {
        method: "POST",
        body: JSON.stringify({
          name: "should-not-exist",
          scope: "write:ingest",
          ttl_days: 1,
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("jwt_required");
    }
  });

  test("revoke token → subsequent use returns invalid_token", async ({ api }) => {
    const r = await api.fetch<CreateTokenResp>("/v1/me/tokens", {
      method: "POST",
      body: JSON.stringify({
        name: "e2e-revoke",
        scope: "read",
        ttl_days: 1,
      }),
    });
    await api.fetch(`/v1/me/tokens/${r.metadata.id}`, { method: "DELETE" });

    const revokedC = createApiClient(r.token);
    try {
      await revokedC.fetch("/v1/me/tokens");
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(401);
      expect(err.code).toBe("invalid_token");
    }
  });

  test("invalid token validation: too-long name", async ({ api }) => {
    try {
      await api.fetch("/v1/me/tokens", {
        method: "POST",
        body: JSON.stringify({
          name: "x".repeat(100),
          scope: "read",
          ttl_days: 1,
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("name_too_long");
    }
  });
});
