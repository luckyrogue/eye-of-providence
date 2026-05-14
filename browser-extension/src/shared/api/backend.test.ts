import { describe, expect, it, vi, beforeEach } from "vitest";
import {
  backendDisplayHost,
  clearConfig,
  dashboardUrlFor,
  getBackend,
  ingest,
  setConfig,
  type EventPayload,
} from "./backend";
describe("backendDisplayHost", () => {
  it("returns host for valid URL", () => {
    expect(backendDisplayHost("https://api.eop.example/api")).toBe("api.eop.example");
  });
  it("returns raw value for garbage URL", () => {
    expect(backendDisplayHost("not-a-url")).toBe("not-a-url");
  });
});
describe("dashboardUrlFor", () => {
  it("strips /api suffix to give dashboard origin", () => {
    expect(dashboardUrlFor("https://eop.example/api")).toBe("https://eop.example");
  });
  it("returns origin when no /api suffix", () => {
    expect(dashboardUrlFor("https://eop.example")).toBe("https://eop.example");
  });
});
describe("setConfig / getBackend / clearConfig", () => {
  beforeEach(async () => {
    await clearConfig();
  });
  it("setConfig writes token + backend, getBackend reads back", async () => {
    await setConfig("tok123", "https://custom.eop.test/api");
    expect(await getBackend()).toBe("https://custom.eop.test/api");
  });
  it("clearConfig removes token", async () => {
    await setConfig("tok123");
    await clearConfig();
    expect(await getBackend()).toBe("https://eop.rysdavletov.org/api");
  });
});
describe("ingest", () => {
  beforeEach(async () => {
    await clearConfig();
    vi.restoreAllMocks();
  });
  const events: EventPayload[] = [
    {
      app_bundle: "chat.openai.com",
      category: "ai",
      source: "browser",
      duration_ms: 1000,
    },
  ];
  it("returns no-token when token missing", async () => {
    const r = await ingest(events);
    expect(r.kind).toBe("no-token");
  });
  it("returns ok on 2xx", async () => {
    await setConfig("tok123");
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () => new Response(JSON.stringify({ accepted: 1, rejected: 0 }), { status: 200 }),
      ),
    );
    const r = await ingest(events);
    expect(r.kind).toBe("ok");
  });
  it("returns retry-later on 5xx", async () => {
    await setConfig("tok123");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("", { status: 503 })),
    );
    const r = await ingest(events);
    expect(r.kind).toBe("retry-later");
  });
  it("returns retry-later on 401", async () => {
    await setConfig("tok123");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("", { status: 401 })),
    );
    const r = await ingest(events);
    expect(r.kind).toBe("retry-later");
  });
  it("returns client-error on 400 (drop batch)", async () => {
    await setConfig("tok123");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("", { status: 400 })),
    );
    const r = await ingest(events);
    expect(r.kind).toBe("client-error");
    if (r.kind === "client-error") expect(r.status).toBe(400);
  });
  it("returns retry-later on network failure", async () => {
    await setConfig("tok123");
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new TypeError("network");
      }),
    );
    const r = await ingest(events);
    expect(r.kind).toBe("retry-later");
  });
});
