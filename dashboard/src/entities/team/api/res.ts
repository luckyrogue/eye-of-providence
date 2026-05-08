import type { Commit, MemberStat, Project, Team, TeamMember } from "./types";

type ListTeamsRes = { teams: Team[] };
type CreateTeamRes = { id: string; name: string; role: string };
type ListMembersRes = { members: TeamMember[] };
type TeamSummaryRes = { members: MemberStat[] };
type ListProjectsRes = { projects: Project[] };
type ListCommitsRes = { commits: Commit[] };
type CreateInviteRes = { code: string; expires_at: string };

export type {
  ListTeamsRes,
  CreateTeamRes,
  ListMembersRes,
  TeamSummaryRes,
  ListProjectsRes,
  ListCommitsRes,
  CreateInviteRes,
};
