import { test, expect } from "../fixtures/index.js";
import { ApiError, createApiClient } from "../helpers/api.js";
import { execSync } from "node:child_process";
function promoteToSuperAdmin(userID: string): void {
  execSync(
    `docker exec eop-postgres psql -U eop -d eop -c "UPDATE users SET global_role='super_admin' WHERE id='${userID}'"`,
    { stdio: "pipe" },
  );
}
test.describe("admin", () => {
  test.beforeAll(async () => {
    const { apiRegister } = await import("../helpers/api.js");
    const { uniqueEmail } = await import("../helpers/db.js");
    try {
      await apiRegister(uniqueEmail("admin-sentinel"), "TestPassword123!");
    } catch {}
  });
  test("regular user gets super_admin_required on /admin/teams", async ({ api }) => {
    try {
      await api.fetch("/v1/admin/teams");
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(403);
      expect(err.code).toBe("super_admin_required");
    }
  });
  test("regular user gets super_admin_required on /admin/users", async ({ api }) => {
    try {
      await api.fetch("/v1/admin/users");
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.code).toBe("super_admin_required");
    }
  });
  test("super_admin can list teams + users + stats", async ({ session }) => {
    promoteToSuperAdmin(session.user_id);
    const c = createApiClient(session.token);
    const teams = await c.fetch<{
      teams: unknown[];
    }>("/v1/admin/teams");
    expect(Array.isArray(teams.teams)).toBe(true);
    const users = await c.fetch<{
      users: unknown[];
    }>("/v1/admin/users");
    expect(Array.isArray(users.users)).toBe(true);
    expect(users.users.length).toBeGreaterThan(0);
    const stats = await c.fetch<{
      users_total: number;
      teams_total: number;
      members_total: number;
    }>("/v1/admin/stats");
    expect(stats.users_total).toBeGreaterThan(0);
  });
  test("super_admin cannot delete themselves (cannot_delete_self)", async ({ session }) => {
    promoteToSuperAdmin(session.user_id);
    const c = createApiClient(session.token);
    try {
      await c.fetch(`/v1/admin/users/${session.user_id}`, { method: "DELETE" });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(409);
      expect(err.code).toBe("cannot_delete_self");
    }
  });
});
