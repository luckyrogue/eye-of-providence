// Минимальный API-клиент. Token берём из localStorage; в Phase 2 заменим на OAuth flow.

const BASE = import.meta.env.VITE_BACKEND_URL || "https://eop-api.rysdavletov.org";

// Кастомное событие, которое App.tsx слушает чтобы корректно разлогинить
// пользователя при 401 от authed эндпоинта (просрочка токена / отозванный).
export const AUTH_FAILED_EVENT = "eop:auth-failed";

function emitAuthFailed() {
  try {
    localStorage.removeItem("eop_token");
    localStorage.removeItem("eop_user_id");
    localStorage.removeItem("eop_display_name");
  } catch {
    // ignore — приватный режим / disabled storage
  }
  window.dispatchEvent(new CustomEvent(AUTH_FAILED_EVENT));
}

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

export type AuthResponse = {
  token: string;
  user_id: string;
  display_name?: string;
  team_id?: string | null;
};

export async function register(email: string, password: string, displayName: string, inviteCode?: string): Promise<AuthResponse> {
  const res = await fetch(`${BASE}/v1/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password, display_name: displayName, invite_code: inviteCode }),
  });
  if (!res.ok) {
    const e = await res.json().catch(() => ({}));
    throw new Error(e.error || `register failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setToken(data.token);
  return data;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${BASE}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const e = await res.json().catch(() => ({}));
    throw new Error(e.error || `login failed: ${res.status}`);
  }
  const data: AuthResponse = await res.json();
  setToken(data.token);
  return data;
}

export type Team = { id: string; name: string; role: string };

export async function listMyTeams(): Promise<Team[]> {
  const res = await authed("/v1/teams");
  if (!res.ok) throw new Error(`teams failed: ${res.status}`);
  const data = await res.json();
  return data.teams ?? [];
}

export async function createTeam(name: string): Promise<{ id: string; name: string; role: string }> {
  const res = await authed("/v1/teams", { method: "POST", body: JSON.stringify({ name }) });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `createTeam failed: ${res.status}`);
  }
  return res.json();
}

export async function updateTeam(teamID: string, name: string): Promise<void> {
  const res = await authed(`/v1/teams/${teamID}`, { method: "PATCH", body: JSON.stringify({ name }) });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `updateTeam failed: ${res.status}`);
  }
}

export async function deleteTeam(teamID: string): Promise<void> {
  const res = await authed(`/v1/teams/${teamID}`, { method: "DELETE" });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `deleteTeam failed: ${res.status}`);
  }
}

export async function updateMemberRole(teamID: string, userID: string, role: string): Promise<void> {
  const res = await authed(`/v1/teams/${teamID}/members/${userID}`, {
    method: "PATCH", body: JSON.stringify({ role }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `updateMemberRole failed: ${res.status}`);
  }
}

export async function removeMember(teamID: string, userID: string): Promise<void> {
  const res = await authed(`/v1/teams/${teamID}/members/${userID}`, { method: "DELETE" });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `removeMember failed: ${res.status}`);
  }
}

export type BetaInfo = { teams_count: number; limit: number; slots_remaining: number };

export async function fetchBetaInfo(): Promise<BetaInfo> {
  const res = await authed("/v1/beta/info");
  if (!res.ok) throw new Error(`betaInfo failed: ${res.status}`);
  return res.json();
}

// --- Me ---

export type Me = {
  user_id: string;
  email: string;
  provider: string;
  display_name?: string;
  github_login?: string;
  global_role?: "user" | "super_admin";
};

export async function fetchMe(): Promise<Me> {
  const res = await authed("/v1/me");
  if (!res.ok) throw new Error(`me failed: ${res.status}`);
  return res.json();
}

// --- Super-admin ---

export type AdminStats = {
  users_total: number;
  teams_total: number;
  members_total: number;
  beta_limit: number;
};

export type AdminTeam = {
  id: string; name: string; plan: string; created_at: string;
  member_count: number;
  owner_email?: string;
};

export type AdminUser = {
  id: string; email: string; display_name: string;
  global_role: string; created_at: string;
  teams_count?: number;
};

export async function adminStats(): Promise<AdminStats> {
  const res = await authed("/v1/admin/stats");
  if (!res.ok) throw new Error(`adminStats failed: ${res.status}`);
  return res.json();
}

export async function adminListTeams(): Promise<AdminTeam[]> {
  const res = await authed("/v1/admin/teams");
  if (!res.ok) throw new Error(`adminListTeams failed: ${res.status}`);
  const data = await res.json();
  return data.teams ?? [];
}

export async function adminListUsers(): Promise<AdminUser[]> {
  const res = await authed("/v1/admin/users");
  if (!res.ok) throw new Error(`adminListUsers failed: ${res.status}`);
  const data = await res.json();
  return data.users ?? [];
}

export async function adminDeleteTeam(teamID: string): Promise<void> {
  const res = await authed(`/v1/admin/teams/${teamID}`, { method: "DELETE" });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `adminDeleteTeam failed: ${res.status}`);
  }
}

export async function adminDeleteUser(userID: string): Promise<void> {
  const res = await authed(`/v1/admin/users/${userID}`, { method: "DELETE" });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `adminDeleteUser failed: ${res.status}`);
  }
}

export async function adminUpdateUser(userID: string, payload: { global_role?: string; display_name?: string }): Promise<void> {
  const res = await authed(`/v1/admin/users/${userID}`, {
    method: "PATCH", body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `adminUpdateUser failed: ${res.status}`);
  }
}

export async function adminAddMember(teamID: string, email: string, role: string): Promise<void> {
  const res = await authed(`/v1/admin/teams/${teamID}/members`, {
    method: "POST", body: JSON.stringify({ email, role }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `adminAddMember failed: ${res.status}`);
  }
}

export type Project = { id: string; name: string; repo_url: string | null; lang_primary: string | null; created_at: string };

export async function listProjects(teamID: string): Promise<Project[]> {
  const res = await authed(`/v1/teams/${teamID}/projects`);
  if (!res.ok) throw new Error(`projects failed: ${res.status}`);
  const data = await res.json();
  return data.projects ?? [];
}

export async function createProject(teamID: string, name: string, repoURL: string): Promise<Project> {
  const res = await authed(`/v1/teams/${teamID}/projects`, {
    method: "POST", body: JSON.stringify({ name, repo_url: repoURL }),
  });
  if (!res.ok) throw new Error(`createProject failed: ${res.status}`);
  return res.json();
}

export type Commit = {
  id: string;
  project_id: string | null;
  user_id: string;
  author: string;
  sha: string;
  message: string;
  branch: string;
  files_changed: number;
  lines_added: number;
  lines_removed: number;
  ai_lines_pct: number | null;
  authored_at: string;
};

export async function listTeamCommits(teamID: string): Promise<Commit[]> {
  const res = await authed(`/v1/teams/${teamID}/commits`);
  if (!res.ok) throw new Error(`commits failed: ${res.status}`);
  const data = await res.json();
  return data.commits ?? [];
}

export type TeamMember = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  created_at: string;
};

export async function listMembers(teamID: string): Promise<TeamMember[]> {
  const res = await authed(`/v1/teams/${teamID}/members`);
  if (!res.ok) throw new Error(`members failed: ${res.status}`);
  const data = await res.json();
  return data.members ?? [];
}

export type MemberStat = {
  id: string;
  display_name: string;
  ai_ms: number;
  manual_ms: number;
  total_ms: number;
  ai_ratio: number;
};

export async function teamSummary(teamID: string): Promise<MemberStat[]> {
  const res = await authed(`/v1/teams/${teamID}/summary`);
  if (!res.ok) throw new Error(`summary failed: ${res.status}`);
  const data = await res.json();
  return data.members ?? [];
}

export async function createInvite(teamID: string): Promise<{ code: string; expires_at: string }> {
  const res = await authed(`/v1/teams/${teamID}/invites`, { method: "POST" });
  if (!res.ok) throw new Error(`invite failed: ${res.status}`);
  return res.json();
}

export type InvitePreview = {
  valid: boolean;
  team_id: string;
  team_name: string;
  uses_left: number;
  expires_at: string | null;
};

export async function previewInvite(code: string): Promise<InvitePreview> {
  const res = await fetch(`${BASE}/v1/invites/${code}`);
  if (!res.ok) throw new Error(`invalid invite`);
  return res.json();
}

export async function acceptInvite(code: string): Promise<void> {
  const res = await authed(`/v1/invites/${code}/accept`, { method: "POST" });
  if (!res.ok) throw new Error(`accept failed: ${res.status}`);
}

export type AuthConfig = { invite_only: boolean; is_first_user: boolean };

export async function fetchAuthConfig(): Promise<AuthConfig> {
  const res = await fetch(`${BASE}/v1/auth/config`);
  if (!res.ok) return { invite_only: false, is_first_user: false };
  return res.json();
}

async function authed(path: string, init: RequestInit = {}): Promise<Response> {
  const token = getToken();
  if (!token) {
    emitAuthFailed();
    throw new Error("not authenticated");
  }
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      ...(init.headers ?? {}),
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  });
  if (res.status === 401) {
    emitAuthFailed();
  }
  return res;
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

export type LangCell = { lang: string; category: string; chars: number; ms: number };

export async function fetchLanguages(days = 30): Promise<LangCell[]> {
  const res = await authed(`/v1/summary/languages?days=${days}`);
  if (!res.ok) throw new Error(`languages failed: ${res.status}`);
  const data = await res.json();
  return data.cells ?? [];
}

export type HeatmapCell = { dow: number; hour: number; category: string; ms: number };

export async function fetchHeatmap(days = 30, tz?: string): Promise<HeatmapCell[]> {
  const q = tz ? `&tz=${encodeURIComponent(tz)}` : "";
  const res = await authed(`/v1/heatmap?days=${days}${q}`);
  if (!res.ok) throw new Error(`heatmap failed: ${res.status}`);
  const data = await res.json();
  return data.cells ?? [];
}

export type TrendPoint = { date: string; category: string; chars: number; ms: number };

export async function fetchTrend(days = 30, tz?: string): Promise<TrendPoint[]> {
  const q = tz ? `&tz=${encodeURIComponent(tz)}` : "";
  const res = await authed(`/v1/trend?days=${days}${q}`);
  if (!res.ok) throw new Error(`trend failed: ${res.status}`);
  const data = await res.json();
  return data.points ?? [];
}

export async function fetchCost(): Promise<Record<string, number>> {
  const res = await authed("/v1/admin/cost");
  if (!res.ok) throw new Error(`cost failed: ${res.status}`);
  return res.json();
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

export type Profile = {
  user_id: string;
  email?: string;
  provider?: string;
  github_login?: string;
};

export async function fetchProfile(): Promise<Profile> {
  const res = await authed("/v1/me/");
  if (!res.ok) throw new Error(`profile failed: ${res.status}`);
  return res.json();
}

export async function deleteMyData(): Promise<void> {
  const res = await authed("/v1/me/data", { method: "DELETE" });
  if (!res.ok) throw new Error(`delete failed: ${res.status}`);
  localStorage.removeItem("eop_token");
  localStorage.removeItem("eop_user_id");
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
