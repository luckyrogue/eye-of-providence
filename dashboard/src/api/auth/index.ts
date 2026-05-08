import { http } from "../../lib/http";
import type { LoginReq, RegisterReq } from "./req";
import type { AuthRes, DevTokenRes } from "./res";
import type { AuthConfig, AuthResponse } from "./types";

export type * from "./types";
export type * from "./req";
export type * from "./res";

function setToken(token: string) {
  localStorage.setItem("eop_token", token);
}

export async function register(
  email: string,
  password: string,
  displayName: string,
  inviteCode?: string,
): Promise<AuthResponse> {
  const body: RegisterReq = {
    email,
    password,
    display_name: displayName,
    invite_code: inviteCode,
  };
  const { data } = await http.post<AuthRes>("/v1/auth/register", body);
  setToken(data.token);
  return data;
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const body: LoginReq = { email, password };
  const { data } = await http.post<AuthRes>("/v1/auth/login", body);
  setToken(data.token);
  return data;
}

export async function devLogin(): Promise<string> {
  const { data } = await http.post<DevTokenRes>("/v1/auth/dev-token");
  setToken(data.token);
  return data.user_id;
}

export async function fetchAuthConfig(): Promise<AuthConfig> {
  try {
    const { data } = await http.get<AuthConfig>("/v1/auth/config");
    return data;
  } catch {
    return { invite_only: false, is_first_user: false };
  }
}
