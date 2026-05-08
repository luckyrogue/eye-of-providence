import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Brain, Copy, FolderGit2, GitCommit, Plus, Settings, Sparkles, Trash2, UserMinus, Users } from "lucide-react";
import {
  listMyTeams, createTeam, listMembers, teamSummary,
  createInvite, listProjects, createProject, listTeamCommits,
  updateTeam, deleteTeam, updateMemberRole, removeMember, fetchBetaInfo,
  type Team, type TeamMember, type MemberStat, type Project, type Commit, type BetaInfo,
} from "./api";
import { formatDate } from "./tz";

export function Teams({ tz }: { tz: string }) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [activeTeam, setActiveTeam] = useState<string | null>(localStorage.getItem("eop_team") || null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [beta, setBeta] = useState<BetaInfo | null>(null);

  async function refreshTeams() {
    setError(null);
    try {
      const [t, b] = await Promise.all([listMyTeams(), fetchBetaInfo().catch(() => null)]);
      setTeams(t);
      setBeta(b);
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
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem("eop_team", id);
  }

  function onTeamChanged(updated: { name?: string; deleted?: boolean }) {
    if (updated.deleted) {
      localStorage.removeItem("eop_team");
      setActiveTeam(null);
    }
    refreshTeams();
  }

  const slotsLeft = beta?.slots_remaining ?? -1;
  const betaFull = beta?.limit && beta.limit > 0 && slotsLeft === 0;

  return (
    <div className="space-y-4">
      {error && <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}

      {beta && beta.limit > 0 && (
        <div className="rounded-xl border bg-gradient-to-br from-purple-500/5 to-blue-500/5 p-4 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-lg bg-foreground/5 flex items-center justify-center">
              <Sparkles className="h-5 w-5 text-purple-500" />
            </div>
            <div>
              <div className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground">Beta program</div>
              <div className="text-sm font-medium mt-0.5">
                {betaFull
                  ? `Все ${beta.limit} мест заняты — open seats coming soon`
                  : `${slotsLeft} из ${beta.limit} мест свободно`}
              </div>
            </div>
          </div>
          <div className="font-display text-3xl font-bold tabular-nums tracking-tightest text-muted-foreground">
            {beta.teams_count}<span className="text-muted-foreground/50">/{beta.limit}</span>
          </div>
        </div>
      )}

      <Card className="card-hover">
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 font-display tracking-tight"><Users className="h-4 w-4" /> Мои команды</CardTitle>
            <CardDescription>Можешь состоять в нескольких — переключайся внизу.</CardDescription>
          </div>
          <Button size="sm" onClick={newTeam} disabled={busy || !!betaFull}>
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
                  <span className="ml-2 font-mono text-[10px] uppercase tracking-widest2 opacity-70">{translateRole(t.role)}</span>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {activeTeam && (
        <TeamDetail
          key={activeTeam}
          teamID={activeTeam}
          team={teams.find((t) => t.id === activeTeam)}
          role={teams.find((t) => t.id === activeTeam)?.role || "member"}
          tz={tz}
          onChanged={onTeamChanged}
        />
      )}
    </div>
  );
}

function PlanBadge({ team }: { team: Team }) {
  const plan = team.subscription_plan || "free";
  const until = team.subscription_until ? new Date(team.subscription_until) : null;
  const isActive = plan !== "free" && (!until || until > new Date());
  const colors: Record<string, string> = {
    free: "bg-muted text-muted-foreground border-border",
    pro: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30",
    team: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/30",
    enterprise: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30",
  };
  return (
    <div className="inline-flex items-center gap-2">
      <span className={`inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-mono uppercase tracking-widest2 ${colors[plan] ?? colors.free}`}>
        {plan}
      </span>
      {until && (
        <span className={`text-[10px] font-mono ${isActive ? "text-muted-foreground" : "text-destructive"}`}>
          {isActive ? "до" : "истёк"} {until.toISOString().slice(0, 10)}
        </span>
      )}
    </div>
  );
}

function translateRole(r: string): string {
  return { owner: "владелец", admin: "админ", member: "участник" }[r] ?? r;
}

function TeamDetail({ teamID, team, role, tz, onChanged }: {
  teamID: string;
  team?: Team;
  role: string;
  tz: string;
  onChanged: (u: { name?: string; deleted?: boolean }) => void;
}) {
  const [tab, setTab] = useState<"members" | "projects" | "commits" | "settings">("members");
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
    <Card className="card-hover">
      <CardHeader className="flex-row items-center justify-between">
        <div>
          <div className="flex items-center gap-3 flex-wrap">
            <CardTitle className="font-display tracking-tight">{team?.name || "Команда"}</CardTitle>
            {team && <PlanBadge team={team} />}
          </div>
          <CardDescription>{members.length} участник{plural(members.length)}{team?.subscription_note ? ` · ${team.subscription_note}` : ""}</CardDescription>
        </div>
        <div className="flex gap-1 text-sm flex-wrap justify-end">
          <TabBtn active={tab === "members"} onClick={() => setTab("members")} icon={<Users className="h-3.5 w-3.5" />}>
            Участники
          </TabBtn>
          <TabBtn active={tab === "projects"} onClick={() => setTab("projects")} icon={<FolderGit2 className="h-3.5 w-3.5" />}>
            Проекты
          </TabBtn>
          <TabBtn active={tab === "commits"} onClick={() => setTab("commits")} icon={<GitCommit className="h-3.5 w-3.5" />}>
            Коммиты
          </TabBtn>
          {role === "owner" && (
            <TabBtn active={tab === "settings"} onClick={() => setTab("settings")} icon={<Settings className="h-3.5 w-3.5" />}>
              Настройки
            </TabBtn>
          )}
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
              {members.map((m) => (
                <MemberRow
                  key={m.id}
                  member={m}
                  stat={stats.find((s) => s.id === m.id)}
                  myRole={role}
                  teamID={teamID}
                  onChanged={refresh}
                  onError={setError}
                />
              ))}
            </ul>
          </>
        )}

        {tab === "projects" && <ProjectsTab teamID={teamID} role={role} projects={projects} onChange={refresh} tz={tz} />}

        {tab === "commits" && (
          <CommitsTable commits={commits} tz={tz} />
        )}

        {tab === "settings" && role === "owner" && team && (
          <TeamSettings teamID={teamID} team={team} onChanged={onChanged} onError={setError} />
        )}
      </CardContent>
    </Card>
  );
}

function MemberRow({ member, stat, myRole, teamID, onChanged, onError }: {
  member: TeamMember;
  stat?: MemberStat;
  myRole: string;
  teamID: string;
  onChanged: () => void;
  onError: (msg: string) => void;
}) {
  const [busy, setBusy] = useState(false);
  const canManage = myRole === "owner" || (myRole === "admin" && member.role !== "owner");
  const canChangeRole = myRole === "owner";

  async function changeRole(role: string) {
    if (role === member.role) return;
    setBusy(true);
    try {
      await updateMemberRole(teamID, member.id, role);
      onChanged();
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!confirm(`Удалить ${member.display_name} из команды?`)) return;
    setBusy(true);
    try {
      await removeMember(teamID, member.id);
      onChanged();
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <li className="flex items-center justify-between rounded-md border p-3 hover:bg-muted/30 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <div className="h-9 w-9 rounded-full bg-gradient-to-br from-primary/30 to-primary/10 flex items-center justify-center text-sm font-medium shrink-0">
          {member.display_name.slice(0, 2).toUpperCase()}
        </div>
        <div className="min-w-0">
          <div className="text-sm font-medium truncate">{member.display_name}</div>
          <div className="text-xs text-muted-foreground truncate">{member.email}</div>
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        {stat && stat.total_ms > 0 && (
          <div className="text-right hidden sm:block">
            <div className="flex items-center gap-1 text-sm justify-end">
              <Brain className="h-3.5 w-3.5 text-purple-500" />
              <span className="font-medium tabular-nums">{stat.ai_ratio}%</span>
            </div>
            <div className="text-[10px] text-muted-foreground tabular-nums font-mono">
              {Math.round(stat.total_ms / 60000)} мин · 7д
            </div>
          </div>
        )}
        {canChangeRole ? (
          <select
            value={member.role}
            disabled={busy}
            onChange={(e) => changeRole(e.target.value)}
            className="rounded-md border bg-background px-2 py-1 text-xs font-mono focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="owner">владелец</option>
            <option value="admin">админ</option>
            <option value="member">участник</option>
          </select>
        ) : (
          <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">{translateRole(member.role)}</span>
        )}
        {canManage && (
          <button
            onClick={remove}
            disabled={busy}
            title="Удалить из команды"
            className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
          >
            <UserMinus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>
    </li>
  );
}

function TeamSettings({ teamID, team, onChanged, onError }: {
  teamID: string;
  team: Team;
  onChanged: (u: { name?: string; deleted?: boolean }) => void;
  onError: (msg: string) => void;
}) {
  const [name, setName] = useState(team.name);
  const [busy, setBusy] = useState(false);

  async function save() {
    if (name.trim() === team.name || !name.trim()) return;
    setBusy(true);
    try {
      await updateTeam(teamID, name.trim());
      onChanged({ name: name.trim() });
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  async function destroy() {
    const confirm1 = prompt(`Введи "${team.name}" чтобы подтвердить удаление команды (необратимо)`);
    if (confirm1 !== team.name) return;
    setBusy(true);
    try {
      await deleteTeam(teamID);
      onChanged({ deleted: true });
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <label className="font-mono text-[11px] uppercase tracking-widest2 text-muted-foreground">Название</label>
        <div className="flex gap-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={busy}
            className="flex-1 rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            maxLength={100}
          />
          <Button onClick={save} disabled={busy || name.trim() === team.name || !name.trim()} size="sm">
            Сохранить
          </Button>
        </div>
      </div>

      <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4">
        <div className="font-mono text-[11px] uppercase tracking-widest2 text-destructive mb-1">Danger zone</div>
        <div className="flex items-center justify-between gap-4">
          <div className="text-sm">
            <div className="font-medium">Удалить команду</div>
            <p className="text-xs text-muted-foreground mt-0.5">
              Уничтожит участников, проекты, инвайты и историю коммитов. Необратимо.
            </p>
          </div>
          <Button onClick={destroy} disabled={busy} variant="destructive" size="sm">
            <Trash2 className="h-3.5 w-3.5 mr-1" /> Удалить
          </Button>
        </div>
      </div>
    </div>
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
