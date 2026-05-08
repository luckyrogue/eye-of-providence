// Server response shapes for /v1/auth/* and /v1/me/*.
import type { Me, Profile } from "./types";

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

type MeRes = Me;
type ProfileRes = Profile;

export type { AuthRes, DevTokenRes, MeRes, ProfileRes };
