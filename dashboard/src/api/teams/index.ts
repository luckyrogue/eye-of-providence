import { http } from "../../lib/http";
import type {
  CreateProjectReq,
  CreateTeamReq,
  UpdateMemberRoleReq,
  UpdateTeamReq,
} from "./req";
import type {
  CreateInviteRes,
  CreateTeamRes,
  ListCommitsRes,
  ListMembersRes,
  ListProjectsRes,
  ListTeamsRes,
  TeamSummaryRes,
} from "./res";
import type { BetaInfo, InvitePreview, Project } from "./types";

export type * from "./types";
export type * from "./req";
export type * from "./res";

// --- Teams ---

export const listMyTeams = () =>
  http.get<ListTeamsRes>("/v1/teams").then((r) => r.data.teams ?? []);

export const fetchBetaInfo = () => http.get<BetaInfo>("/v1/beta/info").then((r) => r.data);

export const createTeam = (name: string) => {
  const body: CreateTeamReq = { name };
  return http.post<CreateTeamRes>("/v1/teams", body).then((r) => r.data);
};

export const updateTeam = (teamID: string, name: string) => {
  const body: UpdateTeamReq = { name };
  return http.patch(`/v1/teams/${teamID}`, body).then(() => undefined);
};

export const deleteTeam = (teamID: string) =>
  http.delete(`/v1/teams/${teamID}`).then(() => undefined);

// --- Members ---

export const listMembers = (teamID: string) =>
  http.get<ListMembersRes>(`/v1/teams/${teamID}/members`).then((r) => r.data.members ?? []);

export const teamSummary = (teamID: string) =>
  http.get<TeamSummaryRes>(`/v1/teams/${teamID}/summary`).then((r) => r.data.members ?? []);

export const updateMemberRole = (teamID: string, userID: string, role: string) => {
  const body: UpdateMemberRoleReq = { role };
  return http.patch(`/v1/teams/${teamID}/members/${userID}`, body).then(() => undefined);
};

export const removeMember = (teamID: string, userID: string) =>
  http.delete(`/v1/teams/${teamID}/members/${userID}`).then(() => undefined);

// --- Invites ---

export const createInvite = (teamID: string) =>
  http.post<CreateInviteRes>(`/v1/teams/${teamID}/invites`).then((r) => r.data);

export const previewInvite = (code: string) =>
  http.get<InvitePreview>(`/v1/invites/${code}`).then((r) => r.data);

export const acceptInvite = (code: string) =>
  http.post(`/v1/invites/${code}/accept`).then(() => undefined);

// --- Projects + Commits ---

export const listProjects = (teamID: string) =>
  http.get<ListProjectsRes>(`/v1/teams/${teamID}/projects`).then((r) => r.data.projects ?? []);

export const createProject = (teamID: string, name: string, repoURL: string) => {
  const body: CreateProjectReq = { name, repo_url: repoURL };
  return http.post<Project>(`/v1/teams/${teamID}/projects`, body).then((r) => r.data);
};

export const listTeamCommits = (teamID: string) =>
  http.get<ListCommitsRes>(`/v1/teams/${teamID}/commits`).then((r) => r.data.commits ?? []);
