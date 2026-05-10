// Backend client. Token и backend URL живут в chrome.storage.local.
// ingest() возвращает success-flag — caller сам решает что делать с failed batch'ем
// (re-queue в persistent storage для retry).

export const DEFAULT_BACKEND = "https://eop.rysdavletov.org/api";

// Display-host для popup — чтобы не дублировать строку. Backend URL парсится
// один раз; на ошибке возвращаем сырое значение.
export function backendDisplayHost(url: string = DEFAULT_BACKEND): string {
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}

const REQUEST_TIMEOUT_MS = 15_000;
const DEV_TOKEN_TIMEOUT_MS = 10_000;

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
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), DEV_TOKEN_TIMEOUT_MS);
  try {
    const res = await fetch(`${backend}/v1/auth/dev-token`, {
      method: "POST",
      signal: ctrl.signal,
    });
    if (!res.ok) throw new Error(`dev-token failed: ${res.status}`);
    const data = (await res.json()) as { token: string };
    return data.token;
  } finally {
    clearTimeout(timer);
  }
}

export type IngestResult =
  | { kind: "ok" }
  | { kind: "no-token" } // не настроено — не retry
  | { kind: "client-error"; status: number } // 4xx, кроме 401/429 — drop, batch битый
  | { kind: "retry-later" }; // 5xx, network, 401, 429

// ingest — отправляет batch. Возвращает результат: caller на retry-later должен
// сложить batch в persistent retry queue и попробовать позже.
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

    // 401 — токен истёк / отозван. retry-later чтобы юзер мог обновить токен.
    // 429 — rate limit. retry-later.
    // 5xx — server-side. retry-later.
    if (res.status === 401 || res.status === 429 || res.status >= 500) {
      console.warn("[eop] ingest retry-later", res.status);
      return { kind: "retry-later" };
    }
    // 4xx (400, 413, и т.п.) — данные битые, retry бесполезен. Дроп.
    console.warn("[eop] ingest client-error, dropping batch", res.status);
    return { kind: "client-error", status: res.status };
  } catch (err) {
    // Network error / abort — retry.
    console.warn("[eop] ingest network error", err);
    return { kind: "retry-later" };
  } finally {
    clearTimeout(timer);
  }
}
