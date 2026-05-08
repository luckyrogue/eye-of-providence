// Ответы /v1/auth/*

type AuthRes = {
  token: string;
  user_id: string;
  display_name?: string;
  team_id?: string | null;
};

type DevTokenRes = {
  token: string;
  user_id: string;
};

export type { AuthRes, DevTokenRes };
