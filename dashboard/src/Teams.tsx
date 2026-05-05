import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Brain, Copy, FolderGit2, GitCommit, Plus, Users } from "lucide-react";
import {
  listMyTeams, createTeam, listMembers, teamSummary,
  createInvite, listProjects, createProject, listTeamCommits,
  type Team, type TeamMember, type MemberStat, type Project, type Commit,
} from "./api";
import { formatDate } from "./tz";

export function Teams({ tz }: { tz: string }) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [activeTeam, setActiveTeam] = useState<string | null>(localStorage.getItem("eop_team") || null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function refreshTeams() {
    setError(null);
    try {
      const t = await listMyTeams();
      setTeams(t);
      if (!activeTeam && t.length) {
        setActiveTeam(t[0].id);
        localStorage.setItem("eop_team", t[0].id);
      }
    } catch (e) {
      setError(String(e));
    }
  }

  useEffect(() => { refreshTeams(); }, []);

  async function newTeam() {
    const name = prompt("Название команды");
    if (!name) return;
    setBusy(true);
    try {
      const r = await createTeam(name);
      await refreshTeams();
      setActiveTeam(r.id);
      localStorage.setItem("eop_team", r.id);
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy(false);
    }
  }

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem("eop_team", id);
  }

  return (
    <div className="space-y-4">
      {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-sm text-destructive">{error}</div>}

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2"><Users className="h-4 w-4" /> Мои команды</CardTitle>
            <CardDescription>Можешь состоять в нескольких — переключайся внизу.</CardDescription>
          </div>
          <Button size="sm" onClick={newTeam} disabled={busy}>
            <Plus className="h-3.5 w-3.5 mr-1" /> Новая команда
          </Button>
        </CardHeader>
        <CardContent>
          {teams.length === 0 ? (
            <p className="text-sm text-muted-foreground">У тебя пока нет команд. Создай новую или попроси приглашение.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {teams.map((t) => (
                <button
                  key={t.id}
                  onClick={() => switchTeam(t.id)}
                  className={`rounded-md border px-3 py-1.5 text-sm transition-colors ${
                    activeTeam === t.id
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-card hover:bg-secondary"
                  }`}
                >
                  {t.name}
                  <span className="ml-2 text-[10px] uppercase opacity-70">{translateRole(t.role)}</span>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {activeTeam && <TeamDetail teamID={activeTeam} role={teams.find((t) => t.id === activeTeam)?.role || "member"} tz={tz} />}
    </div>
  );
}

function translateRole(r: string): string {
  return { owner: "владелец", admin: "админ", member: "участник" }[r] ?? r;
}

function TeamDetail({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const [tab, setTab] = useState<"members" | "projects" | "commits">("members");
  const [members, setMembers] = useState<TeamMember[]>([]);
  const [stats, setStats] = useState<MemberStat[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [commits, setCommits] = useState<Commit[]>([]);
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    setError(null);
    try {
      const [m, s, p, c] = await Promise.all([
        listMembers(teamID),
        teamSummary(teamID),
        listProjects(teamID),
        listTeamCommits(teamID),
      ]);
      setMembers(m);
      setStats(s);
      setProjects(p);
      setCommits(c);
    } catch (e) {
      setError(String(e));
    }
  }

  useEffect(() => { refresh(); }, [teamID]);

  async function makeInvite() {
    try {
      const r = await createInvite(teamID);
      setInviteCode(r.code);
    } catch (e) {
      setError(String(e));
    }
  }

  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  function copyInvite() {
    if (inviteUrl) navigator.clipboard.writeText(inviteUrl);
  }

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <div>
          <CardTitle>Команда</CardTitle>
          <CardDescription>{members.length} участник{plural(members.length)}</CardDescription>
        </div>
        <div className="flex gap-1 text-sm">
          <TabBtn active={tab === "members"} onClick={() => setTab("members")} icon={<Users className="h-3.5 w-3.5" />}>
            Участники
          </TabBtn>
          <TabBtn active={tab === "projects"} onClick={() => setTab("projects")} icon={<FolderGit2 className="h-3.5 w-3.5" />}>
            Проекты
          </TabBtn>
          <TabBtn active={tab === "commits"} onClick={() => setTab("commits")} icon={<GitCommit className="h-3.5 w-3.5" />}>
            Коммиты
          </TabBtn>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && <div className="rounded-md border border-destructive bg-destructive/10 p-2 text-sm text-destructive">{error}</div>}

        {tab === "members" && (
          <>
            {(role === "owner" || role === "admin") && (
              <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
                <div className="flex items-center gap-2 text-sm font-medium">
                  <Plus className="h-4 w-4" /> Пригласить участника
                </div>
                {inviteCode ? (
                  <div className="flex items-center gap-2">
                    <input
                      readOnly
                      value={inviteUrl}
                      className="flex-1 rounded-md border bg-background px-2 py-1.5 text-xs font-mono"
                    />
                    <Button size="sm" variant="outline" onClick={copyInvite}>
                      <Copy className="h-3.5 w-3.5 mr-1" /> Скопировать
                    </Button>
                  </div>
                ) : (
                  <Button size="sm" onClick={makeInvite}>Создать ссылку</Button>
                )}
                <p className="text-xs text-muted-foreground">
                  Ссылка работает 7 дней, до 10 регистраций. Отправь её тому, кого хочешь добавить.
                </p>
              </div>
            )}

            <ul className="space-y-2">
              {members.map((m) => {
                const stat = stats.find((s) => s.id === m.id);
                return (
                  <li key={m.id} className="flex items-center justify-between rounded-md border p-3">
                    <div className="flex items-center gap-3">
                      <div className="h-9 w-9 rounded-full bg-gradient-to-br from-primary/30 to-primary/10 flex items-center justify-center text-sm font-medium">
                        {m.display_name.slice(0, 2).toUpperCase()}
                      </div>
                      <div>
                        <div className="text-sm font-medium">{m.display_name}</div>
                        <div className="text-xs text-muted-foreground">{m.email} · {translateRole(m.role)}</div>
                      </div>
                    </div>
                    {stat && stat.total_ms > 0 && (
                      <div className="text-right">
                        <div className="flex items-center gap-1 text-sm">
                          <Brain className="h-3.5 w-3.5 text-purple-500" />
                          <span className="font-medium tabular-nums">{stat.ai_ratio}%</span>
                          <span className="text-muted-foreground">AI</span>
                        </div>
                        <div className="text-xs text-muted-foreground tabular-nums">
                          {Math.round(stat.total_ms / 60000)} мин · 7 дней
                        </div>
                      </div>
                    )}
                  </li>
                );
              })}
            </ul>
          </>
        )}

        {tab === "projects" && <ProjectsTab teamID={teamID} role={role} projects={projects} onChange={refresh} tz={tz} />}

        {tab === "commits" && (
          <CommitsTable commits={commits} tz={tz} />
        )}
      </CardContent>
    </Card>
  );
}

function TabBtn({ active, onClick, icon, children }: { active: boolean; onClick: () => void; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1 rounded-md px-2.5 py-1 transition-colors ${active ? "bg-secondary" : "text-muted-foreground hover:bg-secondary/50"}`}
    >
      {icon} {children}
    </button>
  );
}

function ProjectsTab({ teamID, role, projects, onChange, tz }: {
  teamID: string; role: string; projects: Project[]; onChange: () => void; tz: string;
}) {
  const [busy, setBusy] = useState(false);

  async function add() {
    const name = prompt("Название проекта");
    if (!name) return;
    const repo = prompt("Repo URL (опционально)") || "";
    setBusy(true);
    try {
      await createProject(teamID, name, repo);
      onChange();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-3">
      {(role === "owner" || role === "admin") && (
        <Button size="sm" onClick={add} disabled={busy}>
          <Plus className="h-3.5 w-3.5 mr-1" /> Новый проект
        </Button>
      )}
      {projects.length === 0 ? (
        <p className="text-sm text-muted-foreground">Пока нет проектов. Создай первый, чтобы привязать коммиты.</p>
      ) : (
        <ul className="space-y-2">
          {projects.map((p) => (
            <li key={p.id} className="rounded-md border p-3">
              <div className="font-medium text-sm">{p.name}</div>
              {p.repo_url && <div className="text-xs text-muted-foreground font-mono">{p.repo_url}</div>}
              <div className="text-xs text-muted-foreground mt-1">создан {formatDate(p.created_at, tz)}</div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function CommitsTable({ commits, tz }: { commits: Commit[]; tz: string }) {
  if (commits.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        Пока нет коммитов. Установи git post-commit hook, и коммиты команды будут видны здесь.
      </p>
    );
  }
  return (
    <div className="overflow-x-auto rounded-md border">
      <table className="w-full text-sm">
        <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>
            <th className="py-2.5 px-3 text-left">Время</th>
            <th className="py-2.5 px-3 text-left">Автор</th>
            <th className="py-2.5 px-3 text-left">SHA</th>
            <th className="py-2.5 px-3 text-left">Сообщение</th>
            <th className="py-2.5 px-3 text-right">+/-</th>
            <th className="py-2.5 px-3 text-right">AI %</th>
          </tr>
        </thead>
        <tbody>
          {commits.map((c) => (
            <tr key={c.id} className="border-t hover:bg-muted/30">
              <td className="py-2 px-3 font-mono text-xs whitespace-nowrap">{formatDate(c.authored_at, tz)}</td>
              <td className="py-2 px-3">{c.author}</td>
              <td className="py-2 px-3 font-mono text-xs">{c.sha.slice(0, 7)}</td>
              <td className="py-2 px-3 max-w-md truncate">{c.message}</td>
              <td className="py-2 px-3 text-right tabular-nums text-xs">
                <span className="text-green-600">+{c.lines_added}</span>{" "}
                <span className="text-red-600">-{c.lines_removed}</span>
              </td>
              <td className="py-2 px-3 text-right tabular-nums">
                {c.ai_lines_pct !== null ? `${c.ai_lines_pct}%` : "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function plural(n: number): string {
  if (n % 10 === 1 && n % 100 !== 11) return "";
  if ([2, 3, 4].includes(n % 10) && ![12, 13, 14].includes(n % 100)) return "а";
  return "ов";
}
