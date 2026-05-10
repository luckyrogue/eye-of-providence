// Lightweight REST client для tests. Используется для seed'инга и для
// non-UI операций (e.g. ingest events, отзыв tokens).
//
// Все methods возвращают response JSON или throw'ят `ApiError` с status code.
// Tests могут assert'ить error.code из RFC 7807 ProblemDetails.

const BACKEND_URL = process.env.E2E_BACKEND_URL || "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    public detail: string,
    public body: unknown,
  ) {
    super(`${status} ${code}: ${detail}`);
  }
}

export interface ApiClient {
  baseURL: string;
  token?: string;
  fetch<T = unknown>(path: string, init?: RequestInit): Promise<T>;
  withToken(token: string): ApiClient;
}

export function createApiClient(token?: string): ApiClient {
  return {
    baseURL: BACKEND_URL,
    token,
    async fetch<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
      const headers = new Headers(init.headers);
      headers.set("Content-Type", "application/json");
      if (token) headers.set("Authorization", `Bearer ${token}`);
      headers.set("X-E2E-Test", "1");

      const res = await fetch(`${BACKEND_URL}${path}`, { ...init, headers });
      const text = await res.text();
      let body: unknown;
      try {
        body = text ? JSON.parse(text) : null;
      } catch {
        body = text;
      }
      if (!res.ok) {
        const b = body as { code?: string; detail?: string; error?: string };
        throw new ApiError(
          res.status,
          b?.code || "unknown",
          b?.detail || b?.error || res.statusText,
          body,
        );
      }
      return body as T;
    },
    withToken(t: string): ApiClient {
      return createApiClient(t);
    },
  };
}

// --- Convenience wrappers (typed чтобы tests не плодили inline types) ---

export interface AuthResponse {
  token: string;
  user_id: string;
  display_name?: string;
  team_id?: string | null;
}

export async function apiRegister(
  email: string,
  password: string,
  displayName?: string,
): Promise<AuthResponse> {
  return createApiClient().fetch<AuthResponse>("/v1/auth/register", {
    method: "POST",
    body: JSON.stringify({
      email,
      password,
      display_name: displayName ?? email,
    }),
  });
}

export async function apiLogin(email: string, password: string): Promise<AuthResponse> {
  return createApiClient().fetch<AuthResponse>("/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function apiDevToken(userID?: string): Promise<AuthResponse> {
  const q = userID ? `?user_id=${userID}` : "";
  return createApiClient().fetch<AuthResponse>(`/v1/auth/dev-token${q}`, {
    method: "POST",
  });
}

export interface TeamRow {
  id: string;
  name: string;
  role: string;
  subscription_plan?: string;
}

export async function apiCreateTeam(token: string, name: string): Promise<TeamRow> {
  return createApiClient(token).fetch<TeamRow>("/v1/teams", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export async function apiHealthz(): Promise<{ status: string }> {
  return createApiClient().fetch<{ status: string }>("/healthz");
}
