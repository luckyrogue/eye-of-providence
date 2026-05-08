import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  Avatar,
  Button,
  Card, CardContent, CardDescription, CardHeader, CardTitle,
  DangerZone, EmptyState, Input, PlanBadge, Select, Tab, TabBar,
  useConfirm,
} from "@eop/ui";
import { Brain, Copy, FolderGit2, GitCommit, Plus, Settings, Sparkles, Trash2, UserMinus, Users } from "lucide-react";
import { formatDate } from "./tz";
import { useMutationToast } from "./hooks/useMutationToast";
import {
  useTeams, useBetaInfo, useCreateTeam,
  useMembers, useTeamSummary, useProjects, useTeamCommits,
  useUpdateMemberRole, useRemoveMember, useUpdateTeam, useDeleteTeam,
  useCreateInvite, useCreateProject,
  type Team, type TeamMember, type MemberStat, type Commit,
} from "./api/teams";

export function Teams({ tz }: { tz: string }) {
  const teams = useTeams();
  const beta = useBetaInfo();
  const createTeam = useCreateTeam();
  const runToast = useMutationToast();
  const [activeTeam, setActiveTeam] = useState<string | null>(localStorage.getItem("eop_team"));

  // Если activeTeam ещё не выбран и список загрузился — выбрать первый.
  const teamsList = teams.data ?? [];
  if (teamsList.length > 0 && !activeTeam) {
    const first = teamsList[0].id;
    setActiveTeam(first);
    localStorage.setItem("eop_team", first);
  }
  // Если activeTeam удалён из списка (deleted) — сбросить.
  if (activeTeam && teamsList.length > 0 && !teamsList.find((t) => t.id === activeTeam)) {
    localStorage.removeItem("eop_team");
    setActiveTeam(teamsList[0]?.id ?? null);
  }

  function switchTeam(id: string) {
    setActiveTeam(id);
    localStorage.setItem("eop_team", id);
  }

  async function onNewTeam() {
    const name = prompt("Название команды");
    if (!name?.trim()) return;
    const r = await runToast(createTeam.mutateAsync(name.trim()), {
      success: "Команда создана",
      error: "Не удалось создать команду",
    });
    if (r) switchTeam(r.id);
  }

  const slotsLeft = beta.data?.slots_remaining ?? -1;
  const betaFull = !!(beta.data?.limit && beta.data.limit > 0 && slotsLeft === 0);
  const activeT = teamsList.find((t) => t.id === activeTeam);

  return (
    <div className="space-y-4">
      {beta.data && beta.data.limit > 0 && <BetaBanner beta={beta.data} betaFull={betaFull} />}

      <Card className="card-hover">
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 font-display tracking-tight">
              <Users className="h-4 w-4" /> Мои команды
            </CardTitle>
            <CardDescription>Можешь состоять в нескольких — переключайся внизу.</CardDescription>
          </div>
          <Button size="sm" onClick={onNewTeam} disabled={createTeam.isPending || betaFull}>
            <Plus className="h-3.5 w-3.5 mr-1" /> Новая команда
          </Button>
        </CardHeader>
        <CardContent>
          {teamsList.length === 0 ? (
            <p className="text-sm text-muted-foreground">У тебя пока нет команд. Создай новую или попроси приглашение.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {teamsList.map((t) => (
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

      {activeTeam && activeT && (
        <TeamDetail key={activeTeam} teamID={activeTeam} team={activeT} tz={tz} />
      )}
    </div>
  );
}

function BetaBanner({ beta, betaFull }: { beta: { teams_count: number; limit: number; slots_remaining: number }; betaFull: boolean }) {
  return (
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
              : `${beta.slots_remaining} из ${beta.limit} мест свободно`}
          </div>
        </div>
      </div>
      <div className="font-display text-3xl font-bold tabular-nums tracking-tightest text-muted-foreground">
        {beta.teams_count}<span className="text-muted-foreground/50">/{beta.limit}</span>
      </div>
    </div>
  );
}

function TeamDetail({ teamID, team, tz }: { teamID: string; team: Team; tz: string }) {
  const role = team.role;
  const [tab, setTab] = useState<"members" | "projects" | "commits" | "settings">("members");
  const members = useMembers(teamID);
  const stats = useTeamSummary(teamID);
  const note = team.subscription_note;
  const memberCount = members.data?.length ?? 0;

  return (
    <Card className="card-hover">
      <CardHeader className="flex-row items-center justify-between">
        <div>
          <div className="flex items-center gap-3 flex-wrap">
            <CardTitle className="font-display tracking-tight">{team.name}</CardTitle>
            <PlanBadge plan={team.subscription_plan ?? "free"} until={team.subscription_until} />
          </div>
          <CardDescription>
            {memberCount} участник{plural(memberCount)}{note ? ` · ${note}` : ""}
          </CardDescription>
        </div>
        <TabBar className="justify-end">
          <Tab active={tab === "members"} onClick={() => setTab("members")} icon={<Users className="h-3.5 w-3.5" />}>Участники</Tab>
          <Tab active={tab === "projects"} onClick={() => setTab("projects")} icon={<FolderGit2 className="h-3.5 w-3.5" />}>Проекты</Tab>
          <Tab active={tab === "commits"} onClick={() => setTab("commits")} icon={<GitCommit className="h-3.5 w-3.5" />}>Коммиты</Tab>
          {role === "owner" && (
            <Tab active={tab === "settings"} onClick={() => setTab("settings")} icon={<Settings className="h-3.5 w-3.5" />}>Настройки</Tab>
          )}
        </TabBar>
      </CardHeader>
      <CardContent className="space-y-4">
        {tab === "members" && (
          <MembersTab teamID={teamID} role={role} members={members.data ?? []} stats={stats.data ?? []} />
        )}
        {tab === "projects" && <ProjectsTab teamID={teamID} role={role} tz={tz} />}
        {tab === "commits" && <CommitsTab teamID={teamID} tz={tz} />}
        {tab === "settings" && role === "owner" && <TeamSettingsTab team={team} />}
      </CardContent>
    </Card>
  );
}

function MembersTab({ teamID, role, members, stats }: {
  teamID: string;
  role: string;
  members: TeamMember[];
  stats: MemberStat[];
}) {
  const createInvite = useCreateInvite(teamID);
  const runToast = useMutationToast();
  const [inviteCode, setInviteCode] = useState<string | null>(null);
  const inviteUrl = inviteCode ? `${window.location.origin}/?invite=${inviteCode}` : "";

  async function makeInvite() {
    const r = await runToast(createInvite.mutateAsync(), { error: "Не удалось создать invite" });
    if (r) setInviteCode(r.code);
  }

  function copyInvite() {
    if (!inviteUrl) return;
    navigator.clipboard.writeText(inviteUrl);
    toast.success("Скопировано");
  }

  return (
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
            <Button size="sm" onClick={makeInvite} disabled={createInvite.isPending}>
              Создать ссылку
            </Button>
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
          />
        ))}
      </ul>
    </>
  );
}

function MemberRow({ member, stat, myRole, teamID }: {
  member: TeamMember;
  stat?: MemberStat;
  myRole: string;
  teamID: string;
}) {
  const updateRole = useUpdateMemberRole(teamID);
  const removeM = useRemoveMember(teamID);
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const canManage = myRole === "owner" || (myRole === "admin" && member.role !== "owner");
  const canChangeRole = myRole === "owner";
  const busy = updateRole.isPending || removeM.isPending;

  async function changeRole(role: string) {
    if (role === member.role) return;
    await runToast(updateRole.mutateAsync({ userID: member.id, role }), {
      success: "Роль обновлена",
      error: "Не удалось изменить роль",
    });
  }

  async function remove() {
    const ok = await confirm({
      title: `Удалить ${member.display_name}?`,
      description: "Участник потеряет доступ к этой команде. Историю событий это не затронет.",
      destructive: true,
      confirmText: "Удалить",
    });
    if (!ok) return;
    await runToast(removeM.mutateAsync(member.id), {
      success: "Участник удалён",
      error: "Не удалось удалить",
    });
  }

  return (
    <li className="flex items-center justify-between rounded-md border p-3 hover:bg-muted/30 transition-colors">
      <div className="flex items-center gap-3 min-w-0">
        <Avatar name={member.display_name} />
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
          <Select
            mono
            value={member.role}
            disabled={busy}
            onChange={(e) => changeRole(e.target.value)}
            className="px-2 py-1 text-xs"
          >
            <option value="owner">владелец</option>
            <option value="admin">админ</option>
            <option value="member">участник</option>
          </Select>
        ) : (
          <span className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
            {translateRole(member.role)}
          </span>
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

function ProjectsTab({ teamID, role, tz }: { teamID: string; role: string; tz: string }) {
  const projects = useProjects(teamID);
  const create = useCreateProject(teamID);
  const runToast = useMutationToast();

  async function add() {
    const name = prompt("Название проекта");
    if (!name?.trim()) return;
    const repoURL = prompt("Repo URL (опционально)") || "";
    await runToast(create.mutateAsync({ name: name.trim(), repoURL }), {
      success: "Проект создан",
      error: "Не удалось создать проект",
    });
  }

  const list = projects.data ?? [];

  return (
    <div className="space-y-3">
      {(role === "owner" || role === "admin") && (
        <Button size="sm" onClick={add} disabled={create.isPending}>
          <Plus className="h-3.5 w-3.5 mr-1" /> Новый проект
        </Button>
      )}
      {list.length === 0 ? (
        <EmptyState
          eyebrow="No projects"
          title="Создай первый проект"
          description="Проекты привязывают коммиты и метрики к репозиториям. После создания установи git post-commit hook."
        />
      ) : (
        <ul className="space-y-2">
          {list.map((p) => (
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

function CommitsTab({ teamID, tz }: { teamID: string; tz: string }) {
  const commits = useTeamCommits(teamID);
  const list: Commit[] = commits.data ?? [];

  if (list.length === 0) {
    return (
      <EmptyState
        eyebrow="No commits yet"
        title="Установи git post-commit hook"
        description="После установки коммиты команды будут показаны здесь с разбивкой AI vs manual."
      />
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
          {list.map((c) => (
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

interface RenameForm {
  name: string;
}

function TeamSettingsTab({ team }: { team: Team }) {
  const update = useUpdateTeam(team.id);
  const del = useDeleteTeam(team.id);
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const { register, handleSubmit, formState: { isDirty }, reset } = useForm<RenameForm>({
    defaultValues: { name: team.name },
  });

  async function onSave(values: RenameForm) {
    const name = values.name.trim();
    if (!name || name === team.name) return;
    const ok = await runToast(update.mutateAsync(name), {
      success: "Сохранено",
      error: "Не удалось переименовать",
    });
    if (ok !== null) reset({ name });
  }

  async function destroy() {
    const ok = await confirm({
      title: `Удалить «${team.name}»?`,
      description: "Уничтожит участников, проекты, инвайты и историю коммитов. Необратимо.",
      typeToConfirm: team.name,
      destructive: true,
      confirmText: "Удалить навсегда",
    });
    if (!ok) return;
    await runToast(del.mutateAsync(), {
      success: "Команда удалена",
      error: "Не удалось удалить",
    });
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleSubmit(onSave)} className="space-y-2">
        <Input
          label="Название"
          {...register("name", { required: true, maxLength: 100 })}
          disabled={update.isPending}
        />
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={update.isPending || !isDirty}>
            Сохранить
          </Button>
        </div>
      </form>

      <DangerZone
        title="Удалить команду"
        description="Уничтожит участников, проекты, инвайты и историю коммитов. Необратимо."
        action={
          <Button onClick={destroy} disabled={del.isPending} variant="destructive" size="sm">
            <Trash2 className="h-3.5 w-3.5 mr-1" /> Удалить
          </Button>
        }
      />
    </div>
  );
}

function translateRole(r: string): string {
  return { owner: "владелец", admin: "админ", member: "участник" }[r] ?? r;
}

function plural(n: number): string {
  if (n % 10 === 1 && n % 100 !== 11) return "";
  if ([2, 3, 4].includes(n % 10) && ![12, 13, 14].includes(n % 100)) return "а";
  return "ов";
}
