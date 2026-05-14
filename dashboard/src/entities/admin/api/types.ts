type AdminStats = {
  users_total: number;
  teams_total: number;
  members_total: number;
  beta_limit: number;
};
type AdminTeam = {
  id: string;
  name: string;
  plan: string;
  created_at: string;
  subscription_plan: string;
  subscription_until: string | null;
  subscription_note: string | null;
  member_count: number;
  owner_email?: string;
};
type AdminUser = {
  id: string;
  email: string;
  display_name: string;
  global_role: string;
  created_at: string;
  teams_count?: number;
};
type Payment = {
  id: string;
  amount_cents: number;
  currency: string;
  method: string;
  note: string;
  covers_until: string;
  paid_at: string;
  recorded_by: string;
};
export type { AdminStats, AdminTeam, AdminUser, Payment };
