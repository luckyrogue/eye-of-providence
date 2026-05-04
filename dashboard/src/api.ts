// Минимальный API-клиент. Token берём из localStorage; в Phase 2 заменим на OAuth flow.

const BASE = import.meta.env.VITE_BACKEND_URL ?? "http://localhost:8080";

export type EventRow = {
  ts: string;
  user_id: string;
  app_bundle: string;
  category: string;
  source: string;
  ai_provider?: string;
  ai_channel?: string;
  duration_ms: number;
  chars_in: number;
};

function getToken(): string | null {
  return localStorage.getItem("eop_token");
}

export function setToken(token: string) {
  localStorage.setItem("eop_token", token);
}

export async function devLogin(): Promise<string> {
  const res = await fetch(`${BASE}/v1/auth/dev-token`, { method: "POST" });
  if (!res.ok) throw new Error(`devLogin failed: ${res.status}`);
  const data = await res.json();
  setToken(data.token);
  return data.user_id;
}

async function authed(path: string, init: RequestInit = {}): Promise<Response> {
  const token = getToken();
  if (!token) throw new Error("not authenticated");
  return fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      ...(init.headers ?? {}),
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
}

export async function fetchRecent(limit = 50): Promise<EventRow[]> {
  const res = await authed(`/v1/events/recent?limit=${limit}`);
  if (!res.ok) throw new Error(`recent failed: ${res.status}`);
  const data = await res.json();
  return data.events ?? [];
}

export async function fetchSummary(days = 7): Promise<Record<string, number>> {
  const res = await authed(`/v1/summary/categories?days=${days}`);
  if (!res.ok) throw new Error(`summary failed: ${res.status}`);
  const data = await res.json();
  return data.categories ?? {};
}

export type Report = {
  id: string;
  period: string;
  model: string;
  body_md: string;
  generated_at: string;
  prompt_version: string;
};

export async function generateReport(period: "weekly" | "monthly" | "daily" = "weekly"): Promise<Report> {
  const res = await authed(`/v1/reports/generate?period=${period}`, { method: "POST" });
  if (!res.ok) throw new Error(`generate failed: ${res.status}`);
  return res.json();
}

export async function listReports(): Promise<Report[]> {
  const res = await authed("/v1/reports/");
  if (!res.ok) throw new Error(`list failed: ${res.status}`);
  const data = await res.json();
  return data.reports ?? [];
}

// Для smoke-теста: посылаем симулированные события с дашборда.
export async function ingestDemoEvent(): Promise<{ accepted: number; rejected: number }> {
  const res = await authed("/v1/ingest", {
    method: "POST",
    body: JSON.stringify({
      events: [
        {
          app_bundle: "com.microsoft.VSCode",
          category: "manual",
          source: "os",
          duration_ms: 30000,
          file_lang: "ts",
          chars_in: 120,
        },
        {
          app_bundle: "chatgpt.com",
          category: "ai",
          source: "browser",
          ai_provider: "openai",
          ai_channel: "chat",
          duration_ms: 60000,
        },
      ],
    }),
  });
  if (!res.ok) throw new Error(`ingest failed: ${res.status}`);
  return res.json();
}
