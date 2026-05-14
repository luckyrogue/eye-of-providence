type Team = {
  id: string;
  name: string;
  role: string;
  subscription_plan?: string;
  subscription_until?: string | null;
  subscription_note?: string | null;
};
type TeamMember = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  joined_at: string;
};
type MemberStat = {
  id: string;
  display_name: string;
  ai_ms: number;
  manual_ms: number;
  total_ms: number;
  ai_ratio: number;
};
type Project = {
  id: string;
  name: string;
  repo_url: string | null;
  lang_primary: string | null;
  created_at: string;
};
type Commit = {
  id: string;
  project_id: string | null;
  user_id: string;
  author: string;
  sha: string;
  message: string;
  branch: string;
  files_changed: number;
  lines_added: number;
  lines_removed: number;
  ai_lines_pct: number | null;
  authored_at: string;
};
type InvitePreview = {
  valid: boolean;
  team_id: string;
  team_name: string;
  uses_left: number;
  expires_at: string | null;
};
type BetaInfo = {
  teams_count: number;
  limit: number;
  slots_remaining: number;
};
export type { Team, TeamMember, MemberStat, Project, Commit, InvitePreview, BetaInfo };
