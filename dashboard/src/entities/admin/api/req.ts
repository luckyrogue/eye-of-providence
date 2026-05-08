import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
import type {
  AdminListPaymentsRes,
  AdminListTeamsRes,
  AdminListUsersRes,
  AdminSetSubscriptionRes,
  AdminStatsRes,
} from "./res";

// --- Request payload types ---

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

// --- Query keys ---

export const adminKeys = {
  stats: ["admin.stats"] as const,
  teams: ["admin.teams"] as const,
  users: ["admin.users"] as const,
  payments: (teamID: string) => ["admin.payments", teamID] as const,
};

// --- Fetchers ---

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

// --- Query hooks ---

export const useAdminStats = () =>
  useQuery({ queryKey: adminKeys.stats, queryFn: adminStats });

export const useAdminTeams = () =>
  useQuery({ queryKey: adminKeys.teams, queryFn: adminListTeams });

export const useAdminUsers = () =>
  useQuery({ queryKey: adminKeys.users, queryFn: adminListUsers });

export const useAdminPayments = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? adminKeys.payments(teamID) : ["admin.payments.disabled"],
    queryFn: () => adminListPayments(teamID!),
    enabled: !!teamID,
  });

// --- Mutation hooks ---

export function useAdminDeleteTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (teamID: string) => adminDeleteTeam(teamID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminKeys.teams });
      qc.invalidateQueries({ queryKey: adminKeys.stats });
    },
  });
}

export function useAdminDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) => adminDeleteUser(userID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminKeys.users });
      qc.invalidateQueries({ queryKey: adminKeys.stats });
    },
  });
}

export function useAdminUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userID, payload }: { userID: string; payload: UpdateUserReq }) =>
      adminUpdateUser(userID, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: adminKeys.users }),
  });
}

export function useAdminAddMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, email, role }: { teamID: string; email: string; role: string }) =>
      adminAddMember(teamID, email, role),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: adminKeys.teams });
      qc.invalidateQueries({ queryKey: adminKeys.users });
    },
  });
}

export function useAdminSetSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, payload }: { teamID: string; payload: SetSubscriptionReq }) =>
      adminSetSubscription(teamID, payload),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: adminKeys.teams });
      qc.invalidateQueries({ queryKey: adminKeys.payments(vars.teamID) });
    },
  });
}
