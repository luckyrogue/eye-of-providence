type CreateTeamReq = { name: string };
type UpdateTeamReq = { name: string };
type UpdateMemberRoleReq = { role: string };
type CreateProjectReq = { name: string; repo_url: string };

export type { CreateTeamReq, UpdateTeamReq, UpdateMemberRoleReq, CreateProjectReq };
