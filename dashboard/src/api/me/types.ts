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

export type { Me, Profile };
