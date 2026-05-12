// Domain types: user identity + auth config.

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
  // TODO: remove the optional + default once backend ships /v1/auth/config with providers.
  providers?: OAuthProvider[];
  passkey_enabled?: boolean;
};

// Linked OAuth identity (user_identities row). Passkeys live in a separate
// table — see Passkey below — but `CountAuthFactors` on the backend sums both
// when evaluating last-factor lockout, so the frontend treats them as two
// orthogonal lists.
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

// Frontend резолвит локализованную строку через t(`insights:${key}`, vars).
type Insight = {
  key: string;
  vars?: Record<string, string | number | boolean>;
};

// Plaintext возвращается ровно один раз — при создании (см. CreateAPITokenRes).
// После этого UI имеет только prefix вида "eop_a4f3…".
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
