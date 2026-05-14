import type { AdminStats, AdminTeam, AdminUser, Payment } from "./types";
type AdminStatsRes = AdminStats;
type AdminListTeamsRes = {
  teams: AdminTeam[];
};
type AdminListUsersRes = {
  users: AdminUser[];
};
type AdminListPaymentsRes = {
  payments: Payment[];
};
type AdminSetSubscriptionRes = {
  payment_id: string | null;
};
export type {
  AdminStatsRes,
  AdminListTeamsRes,
  AdminListUsersRes,
  AdminListPaymentsRes,
  AdminSetSubscriptionRes,
};
