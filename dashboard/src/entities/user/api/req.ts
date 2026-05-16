// /v1/auth/* и /v1/me/* — fetcher функции + TanStack hooks.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import { SESSION_KEYS, clearSession } from "../../../shared/lib/session-storage";
import type {
  APIToken,
  AuthConfig,
  AuthResponse,
  CreateAPITokenRes,
  Insight,
  OnboardingStatus,
} from "./types";
import type { AuthRes, DevTokenRes, MeRes, ProfileRes } from "./res";

export type RegisterReq = {
  email: string;
  password: string;
  display_name: string;
  invite_code?: string;
};

export type LoginReq = { email: string; password: string };

export const userKeys = {
  me: ["me"] as const,
  profile: ["me.profile"] as const,
  authConfig: ["auth.config"] as const,
  onboarding: ["me.onboarding"] as const,
  insights: ["me.insights"] as const,
  tokens: ["me.tokens"] as const,
};

function setToken(token: string) {
  localStorage.setItem(SESSION_KEYS.token, token);
}

// fetchSessionHandoff — reads the one-shot HttpOnly cookie set by the backend
// after an OAuth/passkey login redirect lands on `/auth/complete`. The backend
// clears the cookie on read (one-shot), so this call must happen exactly once
// per redirect.
//
// withCredentials: true — cookie lives on the API origin, axios needs the
// flag to send cross-site credentials when dashboard and API are on different
// hosts. The endpoint ignores Authorization headers, so a stale localStorage
// token doesn't poison the handoff.
export async function fetchSessionHandoff(): Promise<{ token: string }> {
  const { data } = await http.get<{ token: string }>("/v1/me/session-handoff", {
    withCredentials: true,
  });
  return data;
}

// completeSessionHandoff — full flow used by the `/auth/complete` route:
// read the cookie-backed JWT, persist it to localStorage, return it.
export async function completeSessionHandoff(): Promise<string> {
  const { token } = await fetchSessionHandoff();
  setToken(token);
  return token;
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
    // TODO: remove default providers once backend ships providers in /v1/auth/config.
    return {
      invite_only: false,
      is_first_user: false,
      providers: ["github"],
      passkey_enabled: false,
    };
  }
}

export const fetchMe = () => http.get<MeRes>("/v1/me").then((r) => r.data);
export const fetchProfile = () => http.get<ProfileRes>("/v1/me/").then((r) => r.data);

export const fetchOnboardingStatus = () =>
  http.get<OnboardingStatus>("/v1/me/onboarding-status").then((r) => r.data);

export const fetchInsights = (tz?: string) =>
  http
    .get<{ insights: Insight[] }>("/v1/me/insights", {
      params: tz ? { tz } : undefined,
    })
    .then((r) => r.data.insights ?? []);

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

export async function changeMyEmail(currentPassword: string, email: string): Promise<void> {
  const { data } = await http.patch<{ token: string; email: string }>("/v1/me/email", {
    current_password: currentPassword,
    email,
  });
  setToken(data.token);
}

export async function changeMyPassword(
  currentPassword: string,
  newPassword: string,
): Promise<void> {
  const { data } = await http.patch<{ token: string }>("/v1/me/password", {
    current_password: currentPassword,
    new_password: newPassword,
  });
  setToken(data.token);
}

export async function changeMyName(displayName: string, lastName: string | null): Promise<void> {
  await http.patch("/v1/me/name", {
    display_name: displayName,
    last_name: lastName,
  });
}

export const fetchTokens = () =>
  http.get<{ tokens: APIToken[] }>("/v1/me/tokens").then((r) => r.data.tokens ?? []);

export const createAPIToken = (name: string, scope: APIToken["scope"], ttlDays: number) =>
  http
    .post<CreateAPITokenRes>("/v1/me/tokens", { name, scope, ttl_days: ttlDays })
    .then((r) => r.data);

export const revokeAPIToken = (id: string) =>
  http.delete(`/v1/me/tokens/${encodeURIComponent(id)}`).then(() => undefined);

export async function deleteMyData(): Promise<void> {
  await http.delete("/v1/me/data");
  clearSession();
}

// exportMyData — GDPR Article 20. Тянет полный JSON-bundle (profile +
// devices + projects + consent + reports + api_tokens (без hashed) +
// events) и триггерит save dialog в браузере. responseType blob нужен
// чтобы axios не парсил большой JSON в строку — рендерим как файл.
export async function exportMyData(): Promise<void> {
  const r = await http.get("/v1/me/export", { responseType: "blob" });
  const blob =
    r.data instanceof Blob
      ? r.data
      : new Blob([JSON.stringify(r.data)], {
          type: "application/json",
        });
  const url = URL.createObjectURL(blob);
  try {
    const a = document.createElement("a");
    a.href = url;
    // Имя файла backend проставляет через Content-Disposition; если
    // браузер по какой-то причине не подхватил — fallback имя ниже.
    const cd =
      typeof r.headers?.["content-disposition"] === "string"
        ? r.headers["content-disposition"]
        : "";
    const m = /filename="?([^";]+)"?/.exec(cd);
    a.download = m?.[1] ?? `eop-export-${Date.now()}.json`;
    document.body.appendChild(a);
    a.click();
    a.remove();
  } finally {
    // Освобождаем blob через короткий таймаут — Safari нужно время на download.
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
}

export const useAuthConfig = () =>
  useQuery({ queryKey: userKeys.authConfig, queryFn: fetchAuthConfig, staleTime: Infinity });

export const useMe = () =>
  useQuery({
    queryKey: userKeys.me,
    queryFn: fetchMe,
    // 1 повторный запрос; если снова 5xx — useAuthRedirect сбрасывает сессию.
    retry: 1,
  });
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

export const useTokens = () => useQuery({ queryKey: userKeys.tokens, queryFn: fetchTokens });

export function useCreateToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      name,
      scope,
      ttlDays,
    }: {
      name: string;
      scope: APIToken["scope"];
      ttlDays: number;
    }) => createAPIToken(name, scope, ttlDays),
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.tokens }),
  });
}

export function useRevokeToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: revokeAPIToken,
    onSuccess: () => qc.invalidateQueries({ queryKey: userKeys.tokens }),
  });
}

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

export function useExportMyData() {
  return useMutation({
    mutationFn: exportMyData,
  });
}

export function useChangeMyEmail() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ currentPassword, email }: { currentPassword: string; email: string }) =>
      changeMyEmail(currentPassword, email),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: userKeys.me });
      void qc.invalidateQueries({ queryKey: userKeys.profile });
    },
  });
}

export function useChangeMyPassword() {
  return useMutation({
    mutationFn: ({
      currentPassword,
      newPassword,
    }: {
      currentPassword: string;
      newPassword: string;
    }) => changeMyPassword(currentPassword, newPassword),
  });
}

export function useChangeMyName() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ displayName, lastName }: { displayName: string; lastName: string | null }) =>
      changeMyName(displayName, lastName),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: userKeys.me });
      void qc.invalidateQueries({ queryKey: userKeys.profile });
    },
  });
}
