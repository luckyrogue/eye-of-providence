// /v1/auth/* и /v1/me/* — fetcher функции + TanStack hooks.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type { AuthConfig, AuthResponse, Insight, OnboardingStatus } from "./types";
import type { AuthRes, DevTokenRes, MeRes, ProfileRes } from "./res";

// --- Request payload types ---

export type RegisterReq = {
  email: string;
  password: string;
  display_name: string;
  invite_code?: string;
};

export type LoginReq = { email: string; password: string };

// --- Query keys ---

export const userKeys = {
  me: ["me"] as const,
  profile: ["me.profile"] as const,
  authConfig: ["auth.config"] as const,
  onboarding: ["me.onboarding"] as const,
  insights: ["me.insights"] as const,
};

// --- Auth fetchers ---

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

// --- Me fetchers ---

export const fetchMe = () => http.get<MeRes>("/v1/me").then((r) => r.data);
export const fetchProfile = () => http.get<ProfileRes>("/v1/me/").then((r) => r.data);

export const fetchOnboardingStatus = () =>
  http.get<OnboardingStatus>("/v1/me/onboarding-status").then((r) => r.data);

export const fetchInsights = (tz?: string) =>
  http.get<{ insights: Insight[] }>("/v1/me/insights", {
    params: tz ? { tz } : undefined,
  }).then((r) => r.data.insights ?? []);

export async function dismissOnboarding(): Promise<void> {
  await http.post("/v1/me/onboarding/dismiss");
}

export async function updateLocale(locale: string): Promise<void> {
  await http.patch("/v1/me/locale", { locale });
}

export async function forgotPassword(email: string): Promise<void> {
  // Backend всегда отвечает 200 (не палит существование email'а),
  // нам остаётся только показать "если такой email есть, мы прислали письмо".
  await http.post("/v1/auth/forgot-password", { email });
}

export async function resetPassword(token: string, password: string): Promise<void> {
  await http.post("/v1/auth/reset-password", { token, password });
}

export async function deleteMyData(): Promise<void> {
  await http.delete("/v1/me/data");
  localStorage.removeItem("eop_token");
  localStorage.removeItem("eop_user_id");
}

// --- Hooks ---

export const useAuthConfig = () =>
  useQuery({ queryKey: userKeys.authConfig, queryFn: fetchAuthConfig, staleTime: Infinity });

export const useMe = () => useQuery({ queryKey: userKeys.me, queryFn: fetchMe });
export const useProfile = () => useQuery({ queryKey: userKeys.profile, queryFn: fetchProfile });

// useOnboardingStatus — wizard'у нужен polling на step 4 (ждём первое событие).
// Caller передаёт refetchInterval только когда реально ждёт; иначе — once-on-mount.
export const useOnboardingStatus = (opts?: { refetchInterval?: number; enabled?: boolean }) =>
  useQuery({
    queryKey: userKeys.onboarding,
    queryFn: fetchOnboardingStatus,
    refetchInterval: opts?.refetchInterval,
    enabled: opts?.enabled ?? true,
    staleTime: 0,
  });

export const useInsights = (tz?: string) =>
  useQuery({
    queryKey: tz ? [...userKeys.insights, tz] : userKeys.insights,
    queryFn: () => fetchInsights(tz),
    staleTime: 5 * 60 * 1000, // 5 минут — narrative-карточки не меняются часто
  });

export function useDismissOnboarding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: dismissOnboarding,
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.onboarding }),
  });
}

export function useLogin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ email, password }: LoginReq) => login(email, password),
    onSuccess: () => qc.invalidateQueries(),
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

export function useDeleteMyData() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteMyData,
    onSuccess: () => qc.clear(),
  });
}
