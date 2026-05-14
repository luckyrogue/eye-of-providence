export const DEFAULT_BACKEND = "https://eop.rysdavletov.org/api";
export function backendDisplayHost(url: string = DEFAULT_BACKEND): string {
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}
export function dashboardUrlFor(backend: string = DEFAULT_BACKEND): string {
  try {
    const u = new URL(backend);
    const path = u.pathname.replace(/\/api\/?$/, "/");
    u.pathname = path === "" ? "/" : path;
    return u.origin + (path === "/" ? "" : "");
  } catch {
    return "https://eop.rysdavletov.org";
  }
}
const REQUEST_TIMEOUT_MS = 15000;
const PAIR_TIMEOUT_MS = 10000;
export type EventPayload = {
  app_bundle: string;
  category: "idle" | "manual" | "ai" | "reading" | "refactor" | "other";
  source: "browser";
  ai_provider?: string;
  ai_channel?: string;
  duration_ms: number;
  chars_in?: number;
};
async function getConfig() {
  const { eop_token, eop_backend } = await chrome.storage.local.get(["eop_token", "eop_backend"]);
  return {
    token: eop_token as string | undefined,
    backend: (eop_backend as string | undefined) ?? DEFAULT_BACKEND,
  };
}
export async function getBackend(): Promise<string> {
  const { eop_backend } = await chrome.storage.local.get(["eop_backend"]);
  return (eop_backend as string | undefined) ?? DEFAULT_BACKEND;
}
export async function setConfig(token: string, backend?: string) {
  await chrome.storage.local.set({ eop_token: token, eop_backend: backend ?? DEFAULT_BACKEND });
}
export async function clearConfig() {
  await chrome.storage.local.remove(["eop_token", "eop_backend"]);
}
export type PairBeginResponse = {
  pair_id: string;
  secret: string;
  code: string;
  expires_in: number;
};
export type PollResponse =
  | {
      status: "pending";
    }
  | {
      status: "expired";
    }
  | {
      status: "claimed";
      token?: string;
      user_id?: string;
      device_name?: string;
    };
export async function pairBegin(backend = DEFAULT_BACKEND): Promise<PairBeginResponse> {
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), PAIR_TIMEOUT_MS);
  try {
    const res = await fetch(`${backend}/v1/devices/pair`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind: "ext", name: "Browser extension" }),
      signal: ctrl.signal,
    });
    if (!res.ok) throw new Error(`pair failed: ${res.status}`);
    return (await res.json()) as PairBeginResponse;
  } finally {
    clearTimeout(timer);
  }
}
export async function pairPoll(
  pairID: string,
  secret: string,
  backend = DEFAULT_BACKEND,
): Promise<PollResponse> {
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), PAIR_TIMEOUT_MS);
  try {
    const res = await fetch(`${backend}/v1/devices/poll`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ pair_id: pairID, secret }),
      signal: ctrl.signal,
    });
    if (res.status === 404) return { status: "expired" };
    if (!res.ok) throw new Error(`poll failed: ${res.status}`);
    return (await res.json()) as PollResponse;
  } finally {
    clearTimeout(timer);
  }
}
export type IngestResult =
  | {
      kind: "ok";
    }
  | {
      kind: "no-token";
    }
  | {
      kind: "client-error";
      status: number;
    }
  | {
      kind: "retry-later";
    };
export async function ingest(events: EventPayload[]): Promise<IngestResult> {
  const { token, backend } = await getConfig();
  if (!token) {
    console.debug("[eop] no token, skipping ingest of", events.length, "events");
    return { kind: "no-token" };
  }
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), REQUEST_TIMEOUT_MS);
  try {
    const res = await fetch(`${backend}/v1/ingest`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ events }),
      signal: ctrl.signal,
    });
    if (res.ok) return { kind: "ok" };
    if (res.status === 401 || res.status === 429 || res.status >= 500) {
      console.warn("[eop] ingest retry-later", res.status);
      return { kind: "retry-later" };
    }
    console.warn("[eop] ingest client-error, dropping batch", res.status);
    return { kind: "client-error", status: res.status };
  } catch (err) {
    console.warn("[eop] ingest network error", err);
    return { kind: "retry-later" };
  } finally {
    clearTimeout(timer);
  }
}
