// Backend client. Token и backend URL живут в chrome.storage.local.

const DEFAULT_BACKEND = "https://eop.rysdavletov.org/api";

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

export async function setConfig(token: string, backend?: string) {
  await chrome.storage.local.set({ eop_token: token, eop_backend: backend ?? DEFAULT_BACKEND });
}

export async function clearConfig() {
  await chrome.storage.local.remove(["eop_token", "eop_backend"]);
}

export async function fetchDevToken(backend = DEFAULT_BACKEND): Promise<string> {
  const res = await fetch(`${backend}/v1/auth/dev-token`, { method: "POST" });
  if (!res.ok) throw new Error(`dev-token failed: ${res.status}`);
  const data = await res.json();
  return data.token;
}

export async function ingest(events: EventPayload[]): Promise<void> {
  const { token, backend } = await getConfig();
  if (!token) {
    console.debug("[eop] no token, skipping ingest of", events.length, "events");
    return;
  }
  try {
    const res = await fetch(`${backend}/v1/ingest`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ events }),
    });
    if (!res.ok) {
      console.warn("[eop] ingest failed", res.status);
    }
  } catch (err) {
    console.warn("[eop] ingest error", err);
  }
}
