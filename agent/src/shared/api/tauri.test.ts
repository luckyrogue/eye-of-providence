import { describe, expect, it, vi, beforeEach } from "vitest";

const { invokeMock } = vi.hoisted(() => ({ invokeMock: vi.fn() }));
vi.mock("@tauri-apps/api/core", () => ({
  invoke: invokeMock,
}));

import { accountInfo, connectionStatus, pairBegin, pendingCount, setPaused } from "./tauri";

describe("tauri shim", () => {
  beforeEach(() => {
    invokeMock.mockReset();
  });

  it("connectionStatus calls 'connection_status'", async () => {
    invokeMock.mockResolvedValue({
      backend: "online",
      local_api: "online",
      local_api_port: 7373,
      paired: true,
    });
    const r = await connectionStatus();
    expect(r.backend).toBe("online");
    expect(invokeMock).toHaveBeenCalledWith("connection_status");
  });

  it("pendingCount calls 'pending_count'", async () => {
    invokeMock.mockResolvedValue(42);
    expect(await pendingCount()).toBe(42);
  });

  it("accountInfo returns parsed payload", async () => {
    invokeMock.mockResolvedValue({
      paired: true,
      user_id: "u1",
      backend_url: "https://x/api",
      token_from_env: false,
    });
    const r = await accountInfo();
    expect(r.paired).toBe(true);
  });

  it("pairBegin forwards command", async () => {
    invokeMock.mockResolvedValue({ pair_id: "p", secret: "s", code: "ABCDEF", expires_in: 600 });
    const r = await pairBegin();
    expect(r.code).toBe("ABCDEF");
    expect(invokeMock).toHaveBeenCalledWith("pair_begin");
  });

  it("setPaused forwards paused arg", async () => {
    invokeMock.mockResolvedValue(undefined);
    await setPaused(true);
    expect(invokeMock).toHaveBeenCalledWith("set_paused", { paused: true });
  });
});
