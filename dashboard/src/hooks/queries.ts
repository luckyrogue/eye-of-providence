// Централизованные react-query hooks. Каждый endpoint один hook —
// компоненты используют их вместо useEffect + useState.
//
// Принципы:
// - keys заранее определены в `qk` чтобы invalidate'ить точечно
// - после mutation вызываем queryClient.invalidateQueries по нужному key
// - если backend возвращает 401, api.ts emit'ит AUTH_FAILED_EVENT —
//   компонент верхнего уровня уже слушает и редиректит на /login

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchMe,
  fetchRecent,
  fetchSummary,
  fetchLanguages,
  fetchHeatmap,
  fetchTrend,
  listReports,
  listMyTeams,
  fetchBetaInfo,
  listMembers,
  teamSummary,
  listProjects,
  listTeamCommits,
  adminStats,
  adminListTeams,
  adminListUsers,
  adminListPayments,
  // mutations
  createTeam,
  updateTeam,
  deleteTeam,
  updateMemberRole,
  removeMember,
  createInvite,
  createProject,
  generateReport,
  ingestDemoEvent,
  adminDeleteTeam,
  adminDeleteUser,
  adminUpdateUser,
  adminAddMember,
  adminSetSubscription,
} from "../api";

export const qk = {
  me: ["me"] as const,
  recent: (limit: number) => ["events.recent", limit] as const,
  summary: (days: number) => ["events.summary", days] as const,
  languages: (days: number) => ["events.languages", days] as const,
  heatmap: (days: number, tz: string) => ["events.heatmap", days, tz] as const,
  trend: (days: number, tz: string) => ["events.trend", days, tz] as const,
  reports: ["reports"] as const,
  teams: ["teams"] as const,
  beta: ["beta.info"] as const,
  members: (teamID: string) => ["teams", teamID, "members"] as const,
  summaryT: (teamID: string) => ["teams", teamID, "summary"] as const,
  projects: (teamID: string) => ["teams", teamID, "projects"] as const,
  commits: (teamID: string) => ["teams", teamID, "commits"] as const,
  admin: {
    stats: ["admin.stats"] as const,
    teams: ["admin.teams"] as const,
    users: ["admin.users"] as const,
    payments: (teamID: string) => ["admin.payments", teamID] as const,
  },
};

// --- Queries ---

export const useMe = () => useQuery({ queryKey: qk.me, queryFn: fetchMe });
export const useTeams = () => useQuery({ queryKey: qk.teams, queryFn: listMyTeams });
export const useBetaInfo = () => useQuery({ queryKey: qk.beta, queryFn: fetchBetaInfo });
export const useReports = () => useQuery({ queryKey: qk.reports, queryFn: listReports });
export const useRecent = (limit = 20) => useQuery({ queryKey: qk.recent(limit), queryFn: () => fetchRecent(limit) });
export const useSummary = (days = 7) => useQuery({ queryKey: qk.summary(days), queryFn: () => fetchSummary(days) });
export const useLanguages = (days = 30) => useQuery({ queryKey: qk.languages(days), queryFn: () => fetchLanguages(days) });
export const useHeatmap = (days: number, tz: string) =>
  useQuery({ queryKey: qk.heatmap(days, tz), queryFn: () => fetchHeatmap(days, tz) });
export const useTrend = (days: number, tz: string) =>
  useQuery({ queryKey: qk.trend(days, tz), queryFn: () => fetchTrend(days, tz) });

export const useMembers = (teamID: string | null) =>
  useQuery({ queryKey: teamID ? qk.members(teamID) : ["members.disabled"], queryFn: () => listMembers(teamID!), enabled: !!teamID });
export const useTeamSummary = (teamID: string | null) =>
  useQuery({ queryKey: teamID ? qk.summaryT(teamID) : ["summary.disabled"], queryFn: () => teamSummary(teamID!), enabled: !!teamID });
export const useProjects = (teamID: string | null) =>
  useQuery({ queryKey: teamID ? qk.projects(teamID) : ["projects.disabled"], queryFn: () => listProjects(teamID!), enabled: !!teamID });
export const useTeamCommits = (teamID: string | null) =>
  useQuery({ queryKey: teamID ? qk.commits(teamID) : ["commits.disabled"], queryFn: () => listTeamCommits(teamID!), enabled: !!teamID });

export const useAdminStats = () => useQuery({ queryKey: qk.admin.stats, queryFn: adminStats });
export const useAdminTeams = () => useQuery({ queryKey: qk.admin.teams, queryFn: adminListTeams });
export const useAdminUsers = () => useQuery({ queryKey: qk.admin.users, queryFn: adminListUsers });
export const useAdminPayments = (teamID: string | null) =>
  useQuery({
    queryKey: teamID ? qk.admin.payments(teamID) : ["payments.disabled"],
    queryFn: () => adminListPayments(teamID!),
    enabled: !!teamID,
  });

// --- Mutations ---

export function useCreateTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => createTeam(name),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.teams });
      qc.invalidateQueries({ queryKey: qk.beta });
    },
  });
}

export function useUpdateTeam(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => updateTeam(teamID, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.teams }),
  });
}

export function useDeleteTeam(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => deleteTeam(teamID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.teams });
      qc.invalidateQueries({ queryKey: qk.beta });
    },
  });
}

export function useUpdateMemberRole(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userID, role }: { userID: string; role: string }) => updateMemberRole(teamID, userID, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.members(teamID) }),
  });
}

export function useRemoveMember(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) => removeMember(teamID, userID),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.members(teamID) }),
  });
}

export function useCreateInvite(teamID: string) {
  return useMutation({ mutationFn: () => createInvite(teamID) });
}

export function useCreateProject(teamID: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ name, repoURL }: { name: string; repoURL: string }) => createProject(teamID, name, repoURL),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.projects(teamID) }),
  });
}

export function useGenerateReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (period: "weekly" | "monthly") => generateReport(period),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.reports }),
  });
}

export function useIngestDemo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => ingestDemoEvent(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["events.recent"] });
      qc.invalidateQueries({ queryKey: ["events.summary"] });
    },
  });
}

// --- Admin mutations ---

export function useAdminDeleteTeam() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (teamID: string) => adminDeleteTeam(teamID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.admin.teams });
      qc.invalidateQueries({ queryKey: qk.admin.stats });
    },
  });
}

export function useAdminDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) => adminDeleteUser(userID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.admin.users });
      qc.invalidateQueries({ queryKey: qk.admin.stats });
    },
  });
}

export function useAdminUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ userID, payload }: { userID: string; payload: { global_role?: string; display_name?: string } }) =>
      adminUpdateUser(userID, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: qk.admin.users }),
  });
}

export function useAdminAddMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ teamID, email, role }: { teamID: string; email: string; role: string }) =>
      adminAddMember(teamID, email, role),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.admin.teams });
      qc.invalidateQueries({ queryKey: qk.admin.users });
    },
  });
}

export function useAdminSetSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      teamID,
      payload,
    }: {
      teamID: string;
      payload: Parameters<typeof adminSetSubscription>[1];
    }) => adminSetSubscription(teamID, payload),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: qk.admin.teams });
      qc.invalidateQueries({ queryKey: qk.admin.payments(vars.teamID) });
    },
  });
}
