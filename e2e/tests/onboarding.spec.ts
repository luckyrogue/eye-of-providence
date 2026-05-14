import { test, expect } from "../fixtures/index.js";
import { ApiError } from "../helpers/api.js";
import { uniqueTeamName } from "../helpers/db.js";
interface OnboardingStatus {
  teams_count: number;
  has_event: boolean;
  dismissed: boolean;
}
test.describe("onboarding", () => {
  test("fresh user: teams_count=0, has_event=false, dismissed=false", async ({ api }) => {
    const s = await api.fetch<OnboardingStatus>("/v1/me/onboarding-status");
    expect(s.teams_count).toBe(0);
    expect(s.has_event).toBe(false);
    expect(s.dismissed).toBe(false);
  });
  test("after creating team: teams_count=1", async ({ api }) => {
    await api.fetch("/v1/teams", {
      method: "POST",
      body: JSON.stringify({ name: uniqueTeamName("onb") }),
    });
    const s = await api.fetch<OnboardingStatus>("/v1/me/onboarding-status");
    expect(s.teams_count).toBe(1);
  });
  test("dismiss is idempotent (повторный → status ok)", async ({ api }) => {
    const r1 = await api.fetch<{
      status: string;
    }>("/v1/me/onboarding/dismiss", { method: "POST" });
    expect(r1.status).toBe("ok");
    const r2 = await api.fetch<{
      status: string;
    }>("/v1/me/onboarding/dismiss", { method: "POST" });
    expect(r2.status).toBe("ok");
    const s = await api.fetch<OnboardingStatus>("/v1/me/onboarding-status");
    expect(s.dismissed).toBe(true);
  });
  test("set locale = ru → returns ru", async ({ api }) => {
    const r = await api.fetch<{
      locale: string;
    }>("/v1/me/locale", {
      method: "PATCH",
      body: JSON.stringify({ locale: "ru" }),
    });
    expect(r.locale).toBe("ru");
  });
  test("invalid locale → invalid_locale", async ({ api }) => {
    try {
      await api.fetch("/v1/me/locale", {
        method: "PATCH",
        body: JSON.stringify({ locale: "klingon" }),
      });
      throw new Error("expected to throw");
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(400);
      expect(err.code).toBe("invalid_locale");
    }
  });
});
