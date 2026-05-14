import { test, expect } from "../fixtures/index.js";
import { ApiError, createApiClient, apiRegister } from "../helpers/api.js";
import { uniqueEmail, uniqueTeamName } from "../helpers/db.js";
interface TeamCreate {
  id: string;
}
test.describe("sso admin", () => {
  test("get returns configured=false на чистой команде", async ({ api }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-empty") }),
    });
    const r = await api.fetch<{
      configured: boolean;
    }>(`/v1/teams/${team.id}/sso`);
    expect(r.configured).toBe(false);
  });
  test("save oidc config → get returns it (без client_secret)", async ({ api }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-save") }),
    });
    const config = {
      provider: "oidc",
      enabled: true,
      oidc_issuer: "https://accounts.example.com",
      oidc_client_id: "abc123",
      oidc_client_secret: "shhh-secret-456",
      oidc_scopes: ["openid", "email", "profile"],
      allowed_domains: ["acme.com", "subsidiary.io"],
      jit_provision: true,
      jit_role: "member",
    };
    const saved = await api.fetch<{
      config: Record<string, unknown>;
    }>(`/v1/teams/${team.id}/sso`, {
      method: "PUT",
      body: JSON.stringify(config),
    });
    expect(saved.config.provider).toBe("oidc");
    expect(saved.config.has_client_secret).toBe(true);
    expect(saved.config.oidc_client_secret).toBeUndefined();
    expect(saved.config.allowed_domains).toEqual(["acme.com", "subsidiary.io"]);
    const fetched = await api.fetch<{
      configured: boolean;
      config: Record<string, unknown>;
    }>(`/v1/teams/${team.id}/sso`);
    expect(fetched.configured).toBe(true);
    expect(fetched.config.oidc_issuer).toBe("https://accounts.example.com");
  });
  test("save без client_secret на новую config → missing_client_secret", async ({ api }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-no-secret") }),
    });
    try {
      await api.fetch(`/v1/teams/${team.id}/sso`, {
        method: "PUT",
        body: JSON.stringify({
          provider: "oidc",
          enabled: false,
          oidc_issuer: "https://example.com",
          oidc_client_id: "abc",
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("missing_client_secret");
    }
  });
  test("update keeps existing client_secret if not provided", async ({ api }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-update") }),
    });
    await api.fetch(`/v1/teams/${team.id}/sso`, {
      method: "PUT",
      body: JSON.stringify({
        provider: "oidc",
        enabled: false,
        oidc_issuer: "https://idp.example.com",
        oidc_client_id: "id1",
        oidc_client_secret: "secret-v1",
      }),
    });
    const updated = await api.fetch<{
      config: {
        has_client_secret: boolean;
      };
    }>(`/v1/teams/${team.id}/sso`, {
      method: "PUT",
      body: JSON.stringify({
        provider: "oidc",
        enabled: true,
        oidc_issuer: "https://idp.example.com",
        oidc_client_id: "id1",
      }),
    });
    expect(updated.config.has_client_secret).toBe(true);
  });
  test("non-member cannot read team SSO config", async ({ api, session }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-acl") }),
    });
    void session;
    const outsider = await apiRegister(uniqueEmail("sso-outsider"), "TestPassword123!");
    const outClient = createApiClient(outsider.token);
    try {
      await outClient.fetch(`/v1/teams/${team.id}/sso`);
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("not_member");
    }
  });
  test("delete removes config (subsequent GET → configured=false)", async ({ api }) => {
    const team = await api.fetch<TeamCreate>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("sso-del") }),
    });
    await api.fetch(`/v1/teams/${team.id}/sso`, {
      method: "PUT",
      body: JSON.stringify({
        provider: "oidc",
        enabled: false,
        oidc_issuer: "https://to-delete.example",
        oidc_client_id: "x",
        oidc_client_secret: "y",
      }),
    });
    await api.fetch(`/v1/teams/${team.id}/sso`, { method: "DELETE" });
    const after = await api.fetch<{
      configured: boolean;
    }>(`/v1/teams/${team.id}/sso`);
    expect(after.configured).toBe(false);
  });
});
test.describe("sso public endpoints", () => {
  test("start с non-existent team → sso_not_configured", async () => {
    const c = createApiClient();
    try {
      await c.fetch("/v1/sso/start", {
        method: "POST",
        body: JSON.stringify({
          team_id: "00000000-0000-0000-0000-000000000000",
          return_to: "/dashboard",
        }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(404);
      expect(err.code).toBe("sso_not_configured");
    }
  });
  test("invalid team_id → invalid_team_id", async () => {
    const c = createApiClient();
    try {
      await c.fetch("/v1/sso/start", {
        method: "POST",
        body: JSON.stringify({ team_id: "not-a-uuid" }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe("invalid_team_id");
    }
  });
  test("oidc callback missing params → missing_params", async () => {
    const c = createApiClient();
    try {
      await c.fetch("/v1/sso/oidc/callback");
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe("missing_params");
    }
  });
});
