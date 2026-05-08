import { http } from "../../lib/http";
import type { AddMemberReq, SetSubscriptionReq, UpdateUserReq } from "./req";
import type {
  AdminListPaymentsRes,
  AdminListTeamsRes,
  AdminListUsersRes,
  AdminSetSubscriptionRes,
  AdminStatsRes,
} from "./res";

export type * from "./types";
export type * from "./req";
export type * from "./res";

export const adminStats = () => http.get<AdminStatsRes>("/v1/admin/stats").then((r) => r.data);

export const adminListTeams = () =>
  http.get<AdminListTeamsRes>("/v1/admin/teams").then((r) => r.data.teams ?? []);

export const adminListUsers = () =>
  http.get<AdminListUsersRes>("/v1/admin/users").then((r) => r.data.users ?? []);

export const adminListPayments = (teamID: string) =>
  http.get<AdminListPaymentsRes>(`/v1/admin/teams/${teamID}/payments`).then((r) => r.data.payments ?? []);

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
