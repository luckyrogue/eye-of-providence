// Domain types: user identity + auth config.

type Me = {
  user_id: string;
  email: string;
  provider: string;
  display_name?: string;
  github_login?: string;
  global_role?: "user" | "super_admin";
};

type Profile = {
  user_id: string;
  email?: string;
  provider?: string;
  github_login?: string;
};

type AuthResponse = {
  token: string;
  user_id: string;
  display_name?: string;
  team_id?: string | null;
};

type AuthConfig = { invite_only: boolean; is_first_user: boolean };

type OnboardingStatus = {
  teams_count: number;
  has_event: boolean;
  dismissed: boolean;
};

// Insight — narrative-карточка с i18n key + variables. Frontend резолвит
// локализованную строку через t(`insights:${key}`, vars).
type Insight = {
  key: string;
  vars?: Record<string, string | number | boolean>;
};

// APIToken — public API token (для интеграций). Plaintext возвращается ровно
// раз при создании; UI показывает prefix ("eop_a4f3…") после.
type APIToken = {
  id: string;
  name: string;
  scope: "read" | "write:ingest" | "admin";
  prefix: string;
  created_at: string;
  expires_at?: string | null;
  last_used_at?: string | null;
};

type CreateAPITokenRes = {
  token: string; // plaintext, ровно один раз
  metadata: APIToken;
};

export type { Me, Profile, AuthResponse, AuthConfig, OnboardingStatus, Insight, APIToken, CreateAPITokenRes };
