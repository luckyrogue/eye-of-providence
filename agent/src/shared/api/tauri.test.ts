import { describe, expect, it, vi, beforeEach } from "vitest";
const { invokeMock } = vi.hoisted(() => ({ invokeMock: vi.fn() }));
vi.mock("@tauri-apps/api/core", () => ({
  invoke: invokeMock,
}));
import { accountInfo, pairBegin, pairPoll, pendingCount, setBackendUrl, setPaused } from "./tauri";
describe("tauri shim", () => {
  beforeEach(() => {
    invokeMock.mockReset();
  });
  it("pendingCount calls 'pending_count'", async () => {
    invokeMock.mockResolvedValue(42);
    expect(await pendingCount()).toBe(42);
    expect(invokeMock).toHaveBeenCalledWith("pending_count");
  });
  it("accountInfo returns parsed payload", async () => {
    invokeMock.mockResolvedValue({ paired: true, user_id: "u1", backend_url: "https://x" });
    const r = await accountInfo();
    expect(r.paired).toBe(true);
    expect(r.user_id).toBe("u1");
  });
  it("pairBegin forwards command", async () => {
    invokeMock.mockResolvedValue({ pair_id: "p", secret: "s", code: "ABCDEF", expires_in: 600 });
    const r = await pairBegin();
    expect(r.code).toBe("ABCDEF");
    expect(invokeMock).toHaveBeenCalledWith("pair_begin");
  });
  it("pairPoll forwards args as camelCase", async () => {
    invokeMock.mockResolvedValue({ status: "pending", token: null, user_id: null });
    await pairPoll("p1", "secret1");
    expect(invokeMock).toHaveBeenCalledWith("pair_poll", { pairId: "p1", secret: "secret1" });
  });
  it("setBackendUrl forwards url arg", async () => {
    invokeMock.mockResolvedValue(undefined);
    await setBackendUrl("https://api");
    expect(invokeMock).toHaveBeenCalledWith("set_backend_url", { url: "https://api" });
  });
  it("setPaused forwards paused arg", async () => {
    invokeMock.mockResolvedValue(undefined);
    await setPaused(true);
    expect(invokeMock).toHaveBeenCalledWith("set_paused", { paused: true });
  });
});
