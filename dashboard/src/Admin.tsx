import { useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Building2, Crown, Trash2, UserPlus, Users } from "lucide-react";
import {
  adminStats, adminListTeams, adminListUsers,
  adminDeleteTeam, adminDeleteUser, adminUpdateUser, adminAddMember,
  type AdminStats, type AdminTeam, type AdminUser,
} from "./api";
import { formatDate } from "./tz";

type Tab = "overview" | "teams" | "users";

export function Admin({ tz }: { tz: string }) {
  const [tab, setTab] = useState<Tab>("overview");
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [teams, setTeams] = useState<AdminTeam[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    setError(null);
    try {
      const [s, t, u] = await Promise.all([
        adminStats(),
        adminListTeams(),
        adminListUsers(),
      ]);
      setStats(s);
      setTeams(t);
      setUsers(u);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  }

  useEffect(() => { refresh(); }, []);

  return (
    <div className="space-y-4">
      <div className="flex items-baseline justify-between">
        <div>
          <span className="eyebrow">Super admin</span>
          <h2 className="display-head text-3xl mt-2">Управление платформой</h2>
        </div>
        <div className="flex gap-1 text-sm">
          <TabBtn active={tab === "overview"} onClick={() => setTab("overview")}>Обзор</TabBtn>
          <TabBtn active={tab === "teams"} onClick={() => setTab("teams")}>Компании</TabBtn>
          <TabBtn active={tab === "users"} onClick={() => setTab("users")}>Пользователи</TabBtn>
        </div>
      </div>

      {error && <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">{error}</div>}

      {tab === "overview" && stats && <Overview stats={stats} />}
      {tab === "teams" && <TeamsTable teams={teams} users={users} onChange={refresh} onError={setError} tz={tz} />}
      {tab === "users" && <UsersTable users={users} onChange={refresh} onError={setError} tz={tz} />}
    </div>
  );
}

function TabBtn({ active, onClick, children }: { active: boolean; onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`rounded-md px-3 py-1.5 transition-colors ${active ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-secondary"}`}
    >
      {children}
    </button>
  );
}

function Overview({ stats }: { stats: AdminStats }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile label="Пользователей" value={stats.users_total} icon={<Users className="h-4 w-4" />} />
      <StatTile label="Компаний" value={stats.teams_total} hint={stats.beta_limit > 0 ? `beta limit · ${stats.beta_limit}` : "лимит снят"} icon={<Building2 className="h-4 w-4" />} />
      <StatTile label="Membership-связей" value={stats.members_total} icon={<Crown className="h-4 w-4" />} />
    </div>
  );
}

function StatTile({ label, value, hint, icon }: { label: string; value: number; hint?: string; icon: React.ReactNode }) {
  return (
    <Card className="card-hover">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">{label}</CardTitle>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="font-display text-5xl font-bold tabular-nums tracking-tightest">{value}</div>
        {hint && <p className="text-xs text-muted-foreground mt-2 font-mono">{hint}</p>}
      </CardContent>
    </Card>
  );
}

function TeamsTable({ teams, users, onChange, onError, tz }: {
  teams: AdminTeam[]; users: AdminUser[]; onChange: () => void; onError: (e: string) => void; tz: string;
}) {
  const [adding, setAdding] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<"member" | "admin" | "owner">("member");
  const [busy, setBusy] = useState(false);

  async function destroy(teamID: string, name: string) {
    if (!confirm(`Удалить компанию "${name}" со всеми участниками, проектами, инвайтами? Необратимо.`)) return;
    try { await adminDeleteTeam(teamID); onChange(); }
    catch (e) { onError(e instanceof Error ? e.message : String(e)); }
  }

  async function addMember(teamID: string) {
    if (!email.trim()) return;
    setBusy(true);
    try {
      await adminAddMember(teamID, email.trim(), role);
      setAdding(null); setEmail(""); setRole("member");
      onChange();
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally { setBusy(false); }
  }

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">Все компании</CardTitle>
        <CardDescription>{teams.length} компаний в системе</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto rounded-md border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="py-2.5 px-3 text-left">Название</th>
                <th className="py-2.5 px-3 text-left">Plan</th>
                <th className="py-2.5 px-3 text-right">Members</th>
                <th className="py-2.5 px-3 text-left">Owner</th>
                <th className="py-2.5 px-3 text-left">Создана</th>
                <th className="py-2.5 px-3 text-right"></th>
              </tr>
            </thead>
            <tbody>
              {teams.map((t) => (
                <>
                  <tr key={t.id} className="border-t hover:bg-muted/30">
                    <td className="py-2 px-3 font-medium">{t.name}</td>
                    <td className="py-2 px-3 font-mono text-xs">{t.plan}</td>
                    <td className="py-2 px-3 text-right tabular-nums">{t.member_count}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{t.owner_email ?? "—"}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(t.created_at, tz)}</td>
                    <td className="py-2 px-3 text-right">
                      <div className="flex justify-end gap-1">
                        <button
                          onClick={() => setAdding(adding === t.id ? null : t.id)}
                          title="Добавить участника"
                          className="rounded-md p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                        >
                          <UserPlus className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => destroy(t.id, t.name)}
                          title="Удалить компанию"
                          className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                  {adding === t.id && (
                    <tr className="bg-muted/20">
                      <td colSpan={6} className="px-3 py-3">
                        <div className="flex flex-wrap gap-2 items-center">
                          <input
                            list="users-emails"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder="email@example.com"
                            className="flex-1 min-w-48 rounded-md border bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                          />
                          <datalist id="users-emails">
                            {users.map((u) => <option key={u.id} value={u.email} />)}
                          </datalist>
                          <select
                            value={role}
                            onChange={(e) => setRole(e.target.value as typeof role)}
                            className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
                          >
                            <option value="member">member</option>
                            <option value="admin">admin</option>
                            <option value="owner">owner</option>
                          </select>
                          <Button size="sm" onClick={() => addMember(t.id)} disabled={busy || !email.trim()}>Добавить</Button>
                          <Button size="sm" variant="ghost" onClick={() => { setAdding(null); setEmail(""); }}>Отмена</Button>
                        </div>
                      </td>
                    </tr>
                  )}
                </>
              ))}
              {teams.length === 0 && (
                <tr><td colSpan={6} className="py-8 text-center text-muted-foreground">Пока нет компаний</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}

function UsersTable({ users, onChange, onError, tz }: {
  users: AdminUser[]; onChange: () => void; onError: (e: string) => void; tz: string;
}) {
  async function setRole(uid: string, role: string) {
    try { await adminUpdateUser(uid, { global_role: role }); onChange(); }
    catch (e) { onError(e instanceof Error ? e.message : String(e)); }
  }
  async function destroy(uid: string, email: string) {
    if (!confirm(`Удалить пользователя ${email}? Все его данные будут стерты. Необратимо.`)) return;
    try { await adminDeleteUser(uid); onChange(); }
    catch (e) { onError(e instanceof Error ? e.message : String(e)); }
  }

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">Все пользователи</CardTitle>
        <CardDescription>{users.length} учётных записей</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto rounded-md border">
          <table className="w-full text-sm">
            <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="py-2.5 px-3 text-left">Email</th>
                <th className="py-2.5 px-3 text-left">Имя</th>
                <th className="py-2.5 px-3 text-left">Global role</th>
                <th className="py-2.5 px-3 text-right">Команды</th>
                <th className="py-2.5 px-3 text-left">Создан</th>
                <th className="py-2.5 px-3 text-right"></th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-t hover:bg-muted/30">
                  <td className="py-2 px-3 font-mono text-xs">{u.email}</td>
                  <td className="py-2 px-3">{u.display_name}</td>
                  <td className="py-2 px-3">
                    <select
                      value={u.global_role}
                      onChange={(e) => setRole(u.id, e.target.value)}
                      className="rounded-md border bg-background px-2 py-0.5 text-xs font-mono"
                    >
                      <option value="user">user</option>
                      <option value="super_admin">super_admin</option>
                    </select>
                  </td>
                  <td className="py-2 px-3 text-right tabular-nums">{u.teams_count ?? "—"}</td>
                  <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(u.created_at, tz)}</td>
                  <td className="py-2 px-3 text-right">
                    <button
                      onClick={() => destroy(u.id, u.email)}
                      title="Удалить пользователя"
                      className="rounded-md p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}
