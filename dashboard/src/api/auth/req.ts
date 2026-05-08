// /v1/auth/* — fetcher функции + TanStack hooks для login/register/dev/config.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../lib/http";
import type { AuthRes, DevTokenRes } from "./res";
import type { AuthConfig, AuthResponse } from "./types";

// --- Request payload types ---

export type RegisterReq = {
  email: string;
  password: string;
  display_name: string;
  invite_code?: string;
};

export type LoginReq = { email: string; password: string };

// --- Query keys ---

export const authKeys = {
  config: ["auth.config"] as const,
};

// --- Fetchers ---

function setToken(token: string) {
  localStorage.setItem("eop_token", token);
}

export async function register(
  email: string,
  password: string,
  displayName: string,
  inviteCode?: string,
): Promise<AuthResponse> {
  const body: RegisterReq = { email, password, display_name: displayName, invite_code: inviteCode };
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

// --- Hooks ---

// Конфиг auth не меняется в рамках сессии (invite_only / is_first_user
// рассчитывается серверно по count of users), поэтому ставим Infinity staleTime.
export const useAuthConfig = () =>
  useQuery({ queryKey: authKeys.config, queryFn: fetchAuthConfig, staleTime: Infinity });

export function useLogin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ email, password }: LoginReq) => login(email, password),
    onSuccess: () => {
      // После логина инвалидируем все queries — у предыдущего юзера могли быть
      // кэшированные данные, у нового должны подтянуться свежие.
      qc.invalidateQueries();
    },
  });
}

export function useRegister() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: RegisterReq) =>
      register(req.email, req.password, req.display_name, req.invite_code),
    onSuccess: () => qc.invalidateQueries(),
  });
}
