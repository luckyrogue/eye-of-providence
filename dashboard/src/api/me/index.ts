import { http } from "../../lib/http";
import type { MeRes, ProfileRes } from "./res";

export type * from "./types";
export type * from "./res";

export const fetchMe = () => http.get<MeRes>("/v1/me").then((r) => r.data);

export const fetchProfile = () => http.get<ProfileRes>("/v1/me/").then((r) => r.data);

export async function deleteMyData(): Promise<void> {
  await http.delete("/v1/me/data");
  localStorage.removeItem("eop_token");
  localStorage.removeItem("eop_user_id");
}
