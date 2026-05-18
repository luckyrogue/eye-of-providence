import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type {
  AdminListPaymentsRes,
  AdminListTeamsRes,
  AdminListUsersRes,
  AdminSetSubscriptionRes,
  AdminStatsRes,
} from "./res";

export type UpdateUserReq = { global_role?: string; display_name?: string };

export type AddMemberReq = { email: string; role: string };

export type SubscriptionPaymentReq = {
  amount_cents: number;
  currency: string;
  method: string;
  note: string;
  covers_until: string;
};

export type SetSubscriptionReq = {
  plan?: string;
  until?: string | null;
  note?: string | null;
  payment?: SubscriptionPaymentReq;
};

// Все query keys строятся от общего корня `all`, чтобы инвалидация
// одного ключа уровня сущности (`admin`) очищала все её вариации.
export const adminKeys = {
  all: ["admin"] as const,
  stats: () => [...adminKeys.all, "stats"] as const,
  teams: () => [...adminKeys.all, "teams"] as const,
  users: () => [...adminKeys.all, "users"] as const,
  payments: (teamID: string) => [...adminKeys.all, "payments", teamID] as const,
};

export const adminStats = () => http.get<AdminStatsRes>("/v1/admin/stats").then((r) => r.data);

export const adminListTeams = () =>
  http.get<AdminListTeamsRes>("/v1/admin/teams").then((r) => r.data.teams ?? []);

export const adminListUsers = () =>
  http.get<AdminListUsersRes>("/v1/admin/users").then((r) => r.data.users ?? []);

export const adminListPayments = (teamID: string) =>
  http
    .get<AdminListPaymentsRes>(`/v1/admin/teams/${teamID}/payments`)
    .then((r) => r.data.payments ?? []);

export const adminDeleteTeam = (teamID: string) =>
  http.delete(`/v1/admin/teams/${teamID}`).then(() => undefined);

export const adminDeleteUser = (userID: string) =>
  http.delete(`/v1/admin/users/${userID}`).then(() => undefined);

export const adminUpdateUser = (userID: string, payload: UpdateUserReq) =>
  http.patch(`/v1/admin/users/${userID}`, payload).then(() => undefined);

export const adminAddMember = (teamID: string, email: string, role: string) => {
  const body: AddMemberReq = { email, role };
  return http.post(`/v1/admin/teams/${teamID}/members`, body).then(() => undefined);
};

export const adminSetSubscription = (teamID: string, payload: SetSubscriptionReq) =>
  http
    .patch<AdminSetSubscriptionRes>(`/v1/admin/teams/${teamID}/subscription`, payload)
    .then((r) => r.data);

export const useAdminStats = () => useQuery({ queryKey: adminKeys.stats(), queryFn: adminStats });

export const useAdminTeams = () =>
  useQuery({ queryKey: adminKeys.teams(), queryFn: adminListTeams });

export const useAdminUsers = () =>
  useQuery({ queryKey: adminKeys.users(), queryFn: adminListUsers });

export const useAdminPayments = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? adminKeys.payments(teamID) : [...adminKeys.all, "payments", "disabled"],
    queryFn: () => adminListPayments(teamID!),
    enabled: !!teamID,
  });

// --- Revenue ---

export type RevenuePayment = {
  id: string;
  team_id: string;
  team_name: string;
  amount_cents?: number;
  currency?: string;
  method?: string;
  covers_until: string;
  paid_at: string;
  note?: string;
};

export type AdminRevenue = {
  total_cents: number;
  last_30d_cents: number;
  currency: string;
  paying_teams: number;
  by_plan: Record<string, number>;
  recent: RevenuePayment[];
};

const fetchRevenue = () => http.get<AdminRevenue>("/v1/admin/revenue").then((r) => r.data);

export const useAdminRevenue = () =>
  useQuery({ queryKey: [...adminKeys.all, "revenue"] as const, queryFn: fetchRevenue });

// --- Audit log ---

export type AuditEntry = {
  id: string;
  actor_id?: string;
  actor_email?: string;
  action: string;
  target_type?: string;
  target_id?: string;
  metadata?: Record<string, unknown>;
  ip?: string;
  user_agent?: string;
  created_at: string;
};

export type AuditFilter = {
  action?: string;
  target_type?: string;
  target_id?: string;
  actor_id?: string;
  limit?: number;
};

const fetchAudit = (f: AuditFilter) =>
  http
    .get<{ entries: AuditEntry[]; limit: number; offset: number }>("/v1/admin/audit", {
      params: f,
    })
    .then((r) => r.data.entries ?? []);

export const useAdminAudit = (f: AuditFilter = {}) =>
  useQuery({
    queryKey: [...adminKeys.all, "audit", f] as const,
    queryFn: () => fetchAudit(f),
  });

// --- SSO global view ---

export type AdminSSOConfig = {
  team_id: string;
  team_name: string;
  provider: string;
  enabled: boolean;
  oidc_issuer: string;
  oidc_client_id: string;
  allowed_domains: string[];
  jit_provision: boolean;
  jit_role: string;
  created_at: string;
  updated_at: string;
};

const fetchSSOConfigs = () =>
  http.get<{ configs: AdminSSOConfig[] }>("/v1/admin/sso").then((r) => r.data.configs ?? []);

const adminSSODisable = (teamID: string) =>
  http.post(`/v1/admin/sso/${teamID}/disable`).then(() => undefined);

export const useAdminSSOList = () =>
  useQuery({ queryKey: [...adminKeys.all, "sso"] as const, queryFn: fetchSSOConfigs });

export function useAdminSSODisable() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (teamID: string) => adminSSODisable(teamID),
    onSuccess: () => qc.invalidateQueries({ queryKey: [...adminKeys.all, "sso"] }),
  });
}

export function useAdminDeleteTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (teamID: string) => adminDeleteTeam(teamID),
    // Удаление команды затрагивает все admin-запросы (teams/stats/payments).
    onSuccess: () => qc.invalidateQueries({ queryKey: adminKeys.all }),
  });
}

export function useAdminDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) => adminDeleteUser(userID),
    // Удаление user'а каскадно меняет users/stats и потенциально teams.
    onSuccess: () => qc.invalidateQueries({ queryKey: adminKeys.all }),
  });
}

export function useAdminUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userID, payload }: { userID: string; payload: UpdateUserReq }) =>
      adminUpdateUser(userID, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: adminKeys.users() }),
  });
}

export function useAdminAddMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, email, role }: { teamID: string; email: string; role: string }) =>
      adminAddMember(teamID, email, role),
    onSuccess: () =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminKeys.teams() }),
        qc.invalidateQueries({ queryKey: adminKeys.users() }),
      ]),
  });
}

export function useAdminSetSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, payload }: { teamID: string; payload: SetSubscriptionReq }) =>
      adminSetSubscription(teamID, payload),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminKeys.teams() }),
        qc.invalidateQueries({ queryKey: adminKeys.payments(vars.teamID) }),
      ]),
  });
}

// --- Email templates ---

export type EmailTemplateKey = "password_reset" | "team_invite" | "subscription_activated";
export type EmailTemplateLocale = "en" | "ru" | "kk" | "es";

export type EmailTemplateRow = {
  key: EmailTemplateKey;
  locale: EmailTemplateLocale;
  has_override: boolean;
  updated_at: string | null;
  updated_by: string | null;
  updated_by_email: string | null;
};

export type EmailTemplateDetail = {
  key: EmailTemplateKey;
  locale: EmailTemplateLocale;
  subject: string;
  body_html: string;
  body_text: string;
  is_default: boolean;
  updated_at: string | null;
  updated_by: string | null;
};

export type EmailTemplateSaveReq = {
  subject: string;
  body_html: string;
  body_text: string;
};

const adminEmailTemplatesKey = () => [...adminKeys.all, "email_templates"] as const;
const adminEmailTemplateDetailKey = (k: EmailTemplateKey, l: EmailTemplateLocale) =>
  [...adminKeys.all, "email_templates", k, l] as const;

const fetchEmailTemplates = () =>
  http
    .get<{ templates: EmailTemplateRow[] }>("/v1/admin/email-templates")
    .then((r) => r.data.templates ?? []);

// Backend ждёт `:key/:locale` в path (`teams.RegisterAdminRoutes`) — query
// форма возвращала 404 "Cannot GET /v1/admin/email-templates/<key>".
const emailTemplatePath = (key: EmailTemplateKey, locale: EmailTemplateLocale) =>
  `/v1/admin/email-templates/${encodeURIComponent(key)}/${encodeURIComponent(locale)}`;

const fetchEmailTemplate = (key: EmailTemplateKey, locale: EmailTemplateLocale) =>
  http.get<EmailTemplateDetail>(emailTemplatePath(key, locale)).then((r) => r.data);

const saveEmailTemplate = (
  key: EmailTemplateKey,
  locale: EmailTemplateLocale,
  payload: EmailTemplateSaveReq,
) => http.put<EmailTemplateDetail>(emailTemplatePath(key, locale), payload).then((r) => r.data);

const revertEmailTemplate = (key: EmailTemplateKey, locale: EmailTemplateLocale) =>
  http.delete(emailTemplatePath(key, locale)).then(() => undefined);

export const useEmailTemplates = () =>
  useQuery({ queryKey: adminEmailTemplatesKey(), queryFn: fetchEmailTemplates });

export const useEmailTemplate = (
  key: EmailTemplateKey | null,
  locale: EmailTemplateLocale | null,
) =>
  useQuery({
    queryKey:
      key && locale
        ? adminEmailTemplateDetailKey(key, locale)
        : [...adminKeys.all, "email_templates", "disabled"],
    queryFn: () => fetchEmailTemplate(key!, locale!),
    enabled: !!key && !!locale,
  });

export function useSaveEmailTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      key,
      locale,
      payload,
    }: {
      key: EmailTemplateKey;
      locale: EmailTemplateLocale;
      payload: EmailTemplateSaveReq;
    }) => saveEmailTemplate(key, locale, payload),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminEmailTemplatesKey() }),
        qc.invalidateQueries({
          queryKey: adminEmailTemplateDetailKey(vars.key, vars.locale),
        }),
      ]),
  });
}

export function useRevertEmailTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ key, locale }: { key: EmailTemplateKey; locale: EmailTemplateLocale }) =>
      revertEmailTemplate(key, locale),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminEmailTemplatesKey() }),
        qc.invalidateQueries({
          queryKey: adminEmailTemplateDetailKey(vars.key, vars.locale),
        }),
      ]),
  });
}

// --- Team flags ---

export type TeamFlagsPayload = Record<string, boolean | number | null>;

const updateTeamFlags = (teamID: string, payload: TeamFlagsPayload) =>
  http
    .patch<{
      flags: TeamFlagsPayload;
    }>(`/v1/admin/teams/${encodeURIComponent(teamID)}/flags`, payload)
    .then((r) => r.data);

const adminTeamFlagsKey = (teamID: string) => [...adminKeys.all, "team-flags", teamID] as const;

const fetchTeamFlags = (teamID: string) =>
  http
    .get<{ flags: TeamFlagsPayload }>(`/v1/admin/teams/${encodeURIComponent(teamID)}/flags`)
    .then((r) => r.data.flags ?? {});

export const useTeamFlags = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? adminTeamFlagsKey(teamID) : [...adminKeys.all, "team-flags", "disabled"],
    queryFn: () => fetchTeamFlags(teamID!),
    enabled: !!teamID,
  });

export function useUpdateTeamFlags() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, payload }: { teamID: string; payload: TeamFlagsPayload }) =>
      updateTeamFlags(teamID, payload),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminKeys.teams() }),
        qc.invalidateQueries({ queryKey: adminTeamFlagsKey(vars.teamID) }),
      ]),
  });
}

// --- Plan overrides ---

export type PlanOverridesPayload = Record<string, number | null>;

const adminPlanOverridesKey = (teamID: string) =>
  [...adminKeys.all, "plan-overrides", teamID] as const;

const fetchPlanOverrides = (teamID: string) =>
  http
    .get<{
      overrides: PlanOverridesPayload;
      plan_defaults: Record<string, number | null>;
    }>(`/v1/admin/teams/${encodeURIComponent(teamID)}/plan-overrides`)
    .then((r) => r.data);

const updatePlanOverrides = (teamID: string, payload: PlanOverridesPayload) =>
  http
    .patch<{
      overrides: PlanOverridesPayload;
    }>(`/v1/admin/teams/${encodeURIComponent(teamID)}/plan-overrides`, payload)
    .then((r) => r.data);

export const usePlanOverrides = (teamID: string | null) =>
  useQuery({
    queryKey: teamID
      ? adminPlanOverridesKey(teamID)
      : [...adminKeys.all, "plan-overrides", "disabled"],
    queryFn: () => fetchPlanOverrides(teamID!),
    enabled: !!teamID,
  });

export function useUpdateTeamPlanLimits() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, payload }: { teamID: string; payload: PlanOverridesPayload }) =>
      updatePlanOverrides(teamID, payload),
    onSuccess: (_d, vars) =>
      Promise.all([
        qc.invalidateQueries({ queryKey: adminKeys.teams() }),
        qc.invalidateQueries({ queryKey: adminPlanOverridesKey(vars.teamID) }),
      ]),
  });
}

// --- Cross-team webhooks ---

export type CrossTeamWebhook = {
  id: string;
  team_id: string;
  team_name: string;
  url: string;
  events: string[];
  format: string;
  created_at: string;
};

const fetchCrossWebhooks = () =>
  http
    .get<{ webhooks: CrossTeamWebhook[] }>("/v1/admin/webhooks")
    .then((r) => r.data.webhooks ?? []);

const adminCrossWebhooksKey = () => [...adminKeys.all, "cross-webhooks"] as const;

const revokeCrossWebhook = (teamID: string, webhookID: string) =>
  http
    .delete(
      `/v1/admin/teams/${encodeURIComponent(teamID)}/webhooks/${encodeURIComponent(webhookID)}`,
    )
    .then(() => undefined);

export const useAdminCrossWebhooks = () =>
  useQuery({ queryKey: adminCrossWebhooksKey(), queryFn: fetchCrossWebhooks });

export function useAdminRevokeWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, webhookID }: { teamID: string; webhookID: string }) =>
      revokeCrossWebhook(teamID, webhookID),
    onSuccess: () => qc.invalidateQueries({ queryKey: adminCrossWebhooksKey() }),
  });
}

// --- Cross-user API tokens ---

export type CrossUserToken = {
  id: string;
  user_id: string;
  user_email: string;
  prefix: string;
  name: string;
  scope: string;
  created_at: string;
  last_used_at: string | null;
};

const fetchCrossTokens = () =>
  http.get<{ tokens: CrossUserToken[] }>("/v1/admin/api-tokens").then((r) => r.data.tokens ?? []);

const adminCrossTokensKey = () => [...adminKeys.all, "cross-tokens"] as const;

const revokeCrossToken = (tokenID: string) =>
  http.delete(`/v1/admin/api-tokens/${encodeURIComponent(tokenID)}`).then(() => undefined);

export const useAdminCrossTokens = () =>
  useQuery({ queryKey: adminCrossTokensKey(), queryFn: fetchCrossTokens });

export function useAdminRevokeToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (tokenID: string) => revokeCrossToken(tokenID),
    onSuccess: () => qc.invalidateQueries({ queryKey: adminCrossTokensKey() }),
  });
}
