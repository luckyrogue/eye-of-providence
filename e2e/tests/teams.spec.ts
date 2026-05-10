// Teams flow:
//   - create → user становится owner
//   - list возвращает свою команду
//   - update name → 200
//   - invite link с max_uses=10
//   - принять invite другим юзером → стал member
//   - delete команды → cascade в DB
//
// Все через API client (UI-форма тестируется отдельно в teams-ui spec'е).

import { test, expect } from "../fixtures/index.js";
import {
  apiCreateTeam,
  apiRegister,
  createApiClient,
  ApiError,
  type TeamRow,
} from "../helpers/api.js";
import { uniqueEmail, uniqueTeamName } from "../helpers/db.js";

test.describe("teams", () => {
  test("create team → owner role", async ({ api }) => {
    const name = uniqueTeamName("create");
    const team = await api.fetch<TeamRow>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
    expect(team.id).toMatch(/^[0-9a-f-]{36}$/);
    expect(team.name).toBe(name);
    expect(team.role).toBe("owner");
  });

  test("list my teams includes created team", async ({ api }) => {
    const name = uniqueTeamName("list");
    await api.fetch<TeamRow>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
    const out = await api.fetch<{ teams: TeamRow[] }>("/v1/teams");
    const found = out.teams.find((t) => t.name === name);
    expect(found).toBeDefined();
    expect(found?.role).toBe("owner");
  });

  test("update team name", async ({ api }) => {
    const original = uniqueTeamName("update");
    const updated = uniqueTeamName("renamed");
    const team = await api.fetch<TeamRow>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: original }),
    });
    await api.fetch(`/v1/teams/${team.id}`, {
      method: "PATCH",
      body: JSON.stringify({ name: updated }),
    });
    const out = await api.fetch<{ teams: TeamRow[] }>("/v1/teams");
    const found = out.teams.find((t) => t.id === team.id);
    expect(found?.name).toBe(updated);
  });

  test("invalid team name → typed error", async ({ api }) => {
    try {
      await api.fetch("/v1/teams", {
        method: "POST",
        body: JSON.stringify({ name: "" }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("invalid_team_name");
    }
  });

  test("non-owner cannot rename team", async ({ session, api }) => {
    const team = await apiCreateTeam(session.token, uniqueTeamName("ownerrole"));
    // Создаём второго юзера и пытаемся PATCH чужую команду.
    const other = await apiRegister(uniqueEmail("nonowner"), "TestPassword123!");
    const otherClient = createApiClient(other.token);
    try {
      await otherClient.fetch(`/v1/teams/${team.id}`, {
        method: "PATCH",
        body: JSON.stringify({ name: "hijacked" }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      // Может быть not_member (если не в команде) или owner_required (если в как member).
      expect([403]).toContain(err.status);
      expect(["not_member", "owner_required"]).toContain(err.code);
    }
  });

  test("invite link flow: create → another user accepts → becomes member", async ({
    api,
  }) => {
    const team = await api.fetch<TeamRow>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("invite") }),
    });
    const invite = await api.fetch<{ code: string; max_uses: number }>(
      `/v1/teams/${team.id}/invites`,
      { method: "POST", body: JSON.stringify({}) },
    );
    expect(invite.code).toMatch(/^[0-9a-f]{32}$/);
    expect(invite.max_uses).toBe(10);

    // Создаём второго юзера и принимаем invite.
    const newbie = await apiRegister(
      uniqueEmail("invitee"),
      "TestPassword123!",
    );
    const newbieClient = createApiClient(newbie.token);
    const accepted = await newbieClient.fetch<{ team_id: string }>(
      `/v1/invites/${invite.code}/accept`,
      { method: "POST" },
    );
    expect(accepted.team_id).toBe(team.id);

    // Newbie теперь в team.
    const newbieTeams = await newbieClient.fetch<{ teams: TeamRow[] }>(
      "/v1/teams",
    );
    const found = newbieTeams.teams.find((t) => t.id === team.id);
    expect(found?.role).toBe("member");
  });

  test("invalid invite code → invite_invalid", async ({ api }) => {
    try {
      await api.fetch("/v1/invites/deadbeefdeadbeefdeadbeefdeadbeef/accept", {
        method: "POST",
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(404);
      expect(err.code).toBe("invite_invalid");
    }
  });

  test("delete team → no longer in list", async ({ api }) => {
    const team = await api.fetch<TeamRow>("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("delete") }),
    });
    await api.fetch(`/v1/teams/${team.id}`, { method: "DELETE" });
    const out = await api.fetch<{ teams: TeamRow[] }>("/v1/teams");
    expect(out.teams.find((t) => t.id === team.id)).toBeUndefined();
  });

  test("owner limit: cannot create second team", async ({ session }) => {
    const c = createApiClient(session.token);
    await c.fetch("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("first") }),
    });
    try {
      await c.fetch("/v1/teams", {
        method: "POST",
        body: JSON.stringify({ name: uniqueTeamName("second") }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("owner_limit");
    }
  });
});
