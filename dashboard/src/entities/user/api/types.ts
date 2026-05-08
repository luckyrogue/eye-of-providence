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

export type { Me, Profile, AuthResponse, AuthConfig };
