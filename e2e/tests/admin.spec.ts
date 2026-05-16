// Admin endpoints. Регулярный юзер должен получать super_admin_required.
// Промоут юзера в super_admin делаем напрямую через docker exec psql.
//
// Полный admin-flow (set subscription / list payments / delete user) тестируется
// в отдельном integration suite в backend/. Здесь — минимум для guard'ов.

import { test, expect } from "../fixtures/index.js";
import { ApiError, createApiClient } from "../helpers/api.js";
import { execFileSync } from "node:child_process";

function promoteToSuperAdmin(userID: string): void {
  // INSERT может уже быть выполнен (юзер создан выше через API), нам нужно
  // обновить global_role. Через docker exec psql.
  //
  // execFileSync(array) — без shell-инъекции через userID. SQL-инъекция тоже
  // не вылетит: postgres treats $1 как параметр, но мы используем -c с inline
  // строкой, поэтому ограничиваемся UUID-валидацией.
  if (!/^[0-9a-f-]{36}$/i.test(userID)) {
    throw new Error(`promoteToSuperAdmin: invalid userID ${userID}`);
  }
  execFileSync(
    "docker",
    [
      "exec",
      "eop-dev-postgres",
      "psql",
      "-U",
      "eop",
      "-d",
      "eop",
      "-c",
      `UPDATE users SET global_role='super_admin' WHERE id='${userID}'`,
    ],
    { stdio: "pipe" },
  );
}

test.describe("admin", () => {
  // Bootstrap-sentinel: backend промоутит ПЕРВОГО зарегистрированного юзера в
  // super_admin (auth.go: "first user promoted to super_admin"). Если admin
  // suite запускается первым (alphabetical), test#1 регистрирует и обнаруживает
  // что user — super_admin (тест ждёт regular). Регистрируем sentinel заранее,
  // чтобы "consume" first-user privilege и все последующие юзеры стали
  // регулярными.
  test.beforeAll(async () => {
    const { apiRegister } = await import("../helpers/api.js");
    const { uniqueEmail } = await import("../helpers/db.js");
    try {
      await apiRegister(uniqueEmail("admin-sentinel"), "TestPassword123!");
    } catch {
      // Уже есть user в БД — sentinel не нужен. Идемпотентно.
    }
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

    // tv-revocation: после promote'а старый JWT всё ещё валиден (мы не bump'или
    // token_version). Но в test scope этого достаточно — global_role читается
    // из DB при каждом request, не из JWT.
    const c = createApiClient(session.token);

    const teams = await c.fetch<{ teams: unknown[] }>("/v1/admin/teams");
    expect(Array.isArray(teams.teams)).toBe(true);

    const users = await c.fetch<{ users: unknown[] }>("/v1/admin/users");
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
