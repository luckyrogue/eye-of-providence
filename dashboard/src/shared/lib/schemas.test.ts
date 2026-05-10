// Baseline-тесты для централизованных Zod схем.
// Roadmap: добавлять интеграционные тесты по мере роста coverage.
import { describe, expect, it } from "vitest";
import {
  apiTokenSchema,
  inviteSchema,
  loginSchema,
  registerSchema,
  resetPasswordSchema,
  teamSchema,
  webhookSchema,
} from "./schemas";

describe("loginSchema", () => {
  it("accepts a normal login pair", () => {
    expect(loginSchema.safeParse({ email: "user@example.com", password: "x" }).success).toBe(true);
  });

  it("rejects empty email", () => {
    const r = loginSchema.safeParse({ email: "", password: "x" });
    expect(r.success).toBe(false);
  });

  it("rejects malformed email", () => {
    expect(loginSchema.safeParse({ email: "not-an-email", password: "x" }).success).toBe(false);
  });
});

describe("registerSchema", () => {
  it("requires password >= 8 chars", () => {
    expect(
      registerSchema.safeParse({ email: "u@e.com", password: "short", displayName: "User" })
        .success,
    ).toBe(false);
    expect(
      registerSchema.safeParse({ email: "u@e.com", password: "12345678", displayName: "User" })
        .success,
    ).toBe(true);
  });
});

describe("resetPasswordSchema", () => {
  it("requires matching passwords", () => {
    expect(
      resetPasswordSchema.safeParse({ password: "12345678", confirmPassword: "different1" })
        .success,
    ).toBe(false);
    expect(
      resetPasswordSchema.safeParse({ password: "12345678", confirmPassword: "12345678" }).success,
    ).toBe(true);
  });
});

describe("teamSchema / inviteSchema", () => {
  it("teamSchema rejects empty name", () => {
    expect(teamSchema.safeParse({ name: "" }).success).toBe(false);
  });

  it("inviteSchema requires valid email", () => {
    expect(inviteSchema.safeParse({ email: "u@e.com" }).success).toBe(true);
    expect(inviteSchema.safeParse({ email: "broken" }).success).toBe(false);
  });
});

describe("apiTokenSchema", () => {
  it("rejects ttl out of range", () => {
    expect(apiTokenSchema.safeParse({ name: "t", scope: "read", ttlDays: -1 }).success).toBe(false);
    expect(apiTokenSchema.safeParse({ name: "t", scope: "read", ttlDays: 366 }).success).toBe(
      false,
    );
  });

  it("accepts valid token spec", () => {
    expect(
      apiTokenSchema.safeParse({ name: "t", scope: "write:ingest", ttlDays: 30 }).success,
    ).toBe(true);
  });
});

describe("webhookSchema", () => {
  it("requires at least one event", () => {
    expect(
      webhookSchema.safeParse({
        url: "https://example.com/hook",
        events: [],
        format: "raw",
      }).success,
    ).toBe(false);
  });

  it("accepts a complete webhook spec", () => {
    expect(
      webhookSchema.safeParse({
        url: "https://example.com/hook",
        events: ["commit.ingested"],
        format: "slack",
      }).success,
    ).toBe(true);
  });
});
