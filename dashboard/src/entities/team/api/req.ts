// /v1/teams/* — fetchers + TanStack hooks для всех team-endpoint'ов.

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { http } from "../../../shared/api/http";
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

// --- Request payload types ---

export type CreateTeamReq = { name: string };
export type UpdateTeamReq = { name: string };
export type UpdateMemberRoleReq = { role: string };
export type CreateProjectReq = { name: string; repo_url: string };

// --- Query keys ---

export const teamsKeys = {
  list: ["teams"] as const,
  beta: ["teams.beta"] as const,
  members: (teamID: string) => ["teams", teamID, "members"] as const,
  summary: (teamID: string) => ["teams", teamID, "summary"] as const,
  projects: (teamID: string) => ["teams", teamID, "projects"] as const,
  commits: (teamID: string) => ["teams", teamID, "commits"] as const,
};

// --- Fetchers ---

export const listMyTeams = () =>
  http.get<ListTeamsRes>("/v1/teams").then((r) => r.data.teams ?? []);

export const fetchBetaInfo = () => http.get<BetaInfo>("/v1/beta/info").then((r) => r.data);

export const createTeam = (name: string) => {
  const body: CreateTeamReq = { name };
  return http.post<CreateTeamRes>("/v1/teams", body).then((r) => r.data);
};

// encodeURIComponent на user-provided path params: invite-code приходит из
// URL ?invite=, teamID/userID — из click/URL. CodeQL js/client-side-request-forgery
// flag'ит template-strings без encoding.
const enc = encodeURIComponent;

export const updateTeam = (teamID: string, name: string) => {
  const body: UpdateTeamReq = { name };
  return http.patch(`/v1/teams/${enc(teamID)}`, body).then(() => undefined);
};

export const deleteTeam = (teamID: string) =>
  http.delete(`/v1/teams/${enc(teamID)}`).then(() => undefined);

export const listMembers = (teamID: string) =>
  http.get<ListMembersRes>(`/v1/teams/${enc(teamID)}/members`).then((r) => r.data.members ?? []);

export const teamSummary = (teamID: string) =>
  http.get<TeamSummaryRes>(`/v1/teams/${enc(teamID)}/summary`).then((r) => r.data.members ?? []);

export const updateMemberRole = (teamID: string, userID: string, role: string) => {
  const body: UpdateMemberRoleReq = { role };
  return http.patch(`/v1/teams/${enc(teamID)}/members/${enc(userID)}`, body).then(() => undefined);
};

export const removeMember = (teamID: string, userID: string) =>
  http.delete(`/v1/teams/${enc(teamID)}/members/${enc(userID)}`).then(() => undefined);

export const createInvite = (teamID: string, email?: string) => {
  const body = email ? { email } : undefined;
  return http
    .post<CreateInviteRes>(`/v1/teams/${enc(teamID)}/invites`, body)
    .then((r) => r.data);
};

export const previewInvite = (code: string) =>
  http.get<InvitePreview>(`/v1/invites/${enc(code)}`).then((r) => r.data);

export const acceptInvite = (code: string) =>
  http.post(`/v1/invites/${enc(code)}/accept`).then(() => undefined);

export const listProjects = (teamID: string) =>
  http.get<ListProjectsRes>(`/v1/teams/${enc(teamID)}/projects`).then((r) => r.data.projects ?? []);

export const createProject = (teamID: string, name: string, repoURL: string) => {
  const body: CreateProjectReq = { name, repo_url: repoURL };
  return http.post<Project>(`/v1/teams/${enc(teamID)}/projects`, body).then((r) => r.data);
};

export const listTeamCommits = (teamID: string) =>
  http.get<ListCommitsRes>(`/v1/teams/${enc(teamID)}/commits`).then((r) => r.data.commits ?? []);

// --- Query hooks ---

export const useTeams = () => useQuery({ queryKey: teamsKeys.list, queryFn: listMyTeams });
export const useBetaInfo = () => useQuery({ queryKey: teamsKeys.beta, queryFn: fetchBetaInfo });

export const useMembers = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? teamsKeys.members(teamID) : ["members.disabled"],
    queryFn: () => listMembers(teamID!),
    enabled: !!teamID,
  });

export const useTeamSummary = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? teamsKeys.summary(teamID) : ["summary.disabled"],
    queryFn: () => teamSummary(teamID!),
    enabled: !!teamID,
  });

export const useProjects = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? teamsKeys.projects(teamID) : ["projects.disabled"],
    queryFn: () => listProjects(teamID!),
    enabled: !!teamID,
  });

export const useTeamCommits = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? teamsKeys.commits(teamID) : ["commits.disabled"],
    queryFn: () => listTeamCommits(teamID!),
    enabled: !!teamID,
  });

// --- Mutation hooks ---

export function useCreateTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => createTeam(name),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: teamsKeys.list });
      qc.invalidateQueries({ queryKey: teamsKeys.beta });
    },
  });
}

export function useUpdateTeam(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => updateTeam(teamID, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: teamsKeys.list }),
  });
}

export function useDeleteTeam(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => deleteTeam(teamID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: teamsKeys.list });
      qc.invalidateQueries({ queryKey: teamsKeys.beta });
    },
  });
}

export function useUpdateMemberRole(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userID, role }: { userID: string; role: string }) =>
      updateMemberRole(teamID, userID, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: teamsKeys.members(teamID) }),
  });
}

export function useRemoveMember(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) => removeMember(teamID, userID),
    onSuccess: () => qc.invalidateQueries({ queryKey: teamsKeys.members(teamID) }),
  });
}

export function useCreateInvite(teamID: string) {
  return useMutation({ mutationFn: () => createInvite(teamID) });
}

export function useCreateProject(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ name, repoURL }: { name: string; repoURL: string }) =>
      createProject(teamID, name, repoURL),
    onSuccess: () => qc.invalidateQueries({ queryKey: teamsKeys.projects(teamID) }),
  });
}
