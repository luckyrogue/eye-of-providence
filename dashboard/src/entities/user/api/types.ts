type Me = {
  user_id: string;
  email: string;
  provider: string;
  display_name?: string;
  last_name?: string;
  phone?: string;
  github_login?: string;
  global_role?: "user" | "super_admin";
  has_password?: boolean;
  locale?: string;
  created_at?: string;
};
type Profile = {
  user_id: string;
  email?: string;
  provider?: string;
  github_login?: string;
  has_password?: boolean;
  display_name?: string;
  last_name?: string;
  phone?: string;
  global_role?: "user" | "super_admin";
  locale?: string;
  created_at?: string;
};
type AuthResponse = {
  token: string;
  user_id: string;
  display_name?: string;
  team_id?: string | null;
};
type OAuthProvider = "github" | "google" | "apple";
type AuthConfig = {
  invite_only: boolean;
  is_first_user: boolean;
  providers?: OAuthProvider[];
  passkey_enabled?: boolean;
};
type Identity = {
  id: string;
  provider: OAuthProvider;
  subject: string;
  email?: string;
  created_at: string;
};
type Passkey = {
  id: string;
  nickname: string;
  created_at: string;
  last_used_at?: string | null;
};
type OnboardingStatus = {
  teams_count: number;
  has_event: boolean;
  dismissed: boolean;
};
type Insight = {
  key: string;
  vars?: Record<string, string | number | boolean>;
};
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
  token: string;
  metadata: APIToken;
};
export type {
  Me,
  Profile,
  AuthResponse,
  AuthConfig,
  OAuthProvider,
  Identity,
  Passkey,
  OnboardingStatus,
  Insight,
  APIToken,
  CreateAPITokenRes,
};
