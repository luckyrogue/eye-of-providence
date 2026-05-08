import { Fragment, useEffect, useState } from "react";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Building2, CreditCard, Crown, Trash2, UserPlus, Users, X } from "lucide-react";
import {
  adminStats, adminListTeams, adminListUsers,
  adminDeleteTeam, adminDeleteUser, adminUpdateUser, adminAddMember,
  adminSetSubscription, adminListPayments,
  type AdminStats, type AdminTeam, type AdminUser, type Payment,
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
  const [subTeam, setSubTeam] = useState<AdminTeam | null>(null);

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
    <>
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
                  <th className="py-2.5 px-3 text-left">Подписка</th>
                  <th className="py-2.5 px-3 text-right">Members</th>
                  <th className="py-2.5 px-3 text-left">Owner</th>
                  <th className="py-2.5 px-3 text-left">Создана</th>
                  <th className="py-2.5 px-3 text-right"></th>
                </tr>
              </thead>
              <tbody>
                {teams.map((t) => (
                  <Fragment key={t.id}>
                    <tr className="border-t hover:bg-muted/30">
                      <td className="py-2 px-3 font-medium">{t.name}</td>
                      <td className="py-2 px-3"><PlanPill team={t} /></td>
                      <td className="py-2 px-3 text-right tabular-nums">{t.member_count}</td>
                      <td className="py-2 px-3 text-xs text-muted-foreground">{t.owner_email ?? "—"}</td>
                      <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(t.created_at, tz)}</td>
                      <td className="py-2 px-3 text-right">
                        <div className="flex justify-end gap-1">
                          <button
                            onClick={() => setSubTeam(t)}
                            title="Управление подпиской"
                            className="rounded-md p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                          >
                            <CreditCard className="h-3.5 w-3.5" />
                          </button>
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
                  </Fragment>
                ))}
                {teams.length === 0 && (
                  <tr><td colSpan={6} className="py-8 text-center text-muted-foreground">Пока нет компаний</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {subTeam && (
        <SubscriptionModal
          team={subTeam}
          tz={tz}
          onClose={() => setSubTeam(null)}
          onSaved={() => { setSubTeam(null); onChange(); }}
          onError={onError}
        />
      )}
    </>
  );
}

function PlanPill({ team }: { team: { subscription_plan: string; subscription_until: string | null } }) {
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
    <div className="flex items-center gap-2">
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

function SubscriptionModal({ team, tz, onClose, onSaved, onError }: {
  team: AdminTeam;
  tz: string;
  onClose: () => void;
  onSaved: () => void;
  onError: (e: string) => void;
}) {
  const [plan, setPlan] = useState(team.subscription_plan || "free");
  const [until, setUntil] = useState(team.subscription_until ? team.subscription_until.slice(0, 10) : "");
  const [note, setNote] = useState(team.subscription_note ?? "");
  const [recordPayment, setRecordPayment] = useState(true);
  const [amount, setAmount] = useState("");
  const [currency, setCurrency] = useState("USD");
  const [method, setMethod] = useState("manual_transfer");
  const [paymentNote, setPaymentNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [payments, setPayments] = useState<Payment[]>([]);

  useEffect(() => {
    adminListPayments(team.id).then(setPayments).catch(() => {});
  }, [team.id]);

  function quickExtend(months: number) {
    const base = until ? new Date(until) : new Date();
    if (!until || base < new Date()) base.setTime(Date.now());
    base.setMonth(base.getMonth() + months);
    setUntil(base.toISOString().slice(0, 10));
  }

  async function save() {
    setBusy(true);
    try {
      const payload: Parameters<typeof adminSetSubscription>[1] = {
        plan,
        until: until ? new Date(until + "T23:59:59Z").toISOString() : "",
        note,
      };
      const amt = parseInt(amount, 10);
      if (recordPayment && plan !== "free" && amt > 0 && until) {
        payload.payment = {
          amount_cents: amt,
          currency,
          method,
          note: paymentNote,
          covers_until: new Date(until + "T23:59:59Z").toISOString(),
        };
      }
      await adminSetSubscription(team.id, payload);
      onSaved();
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  async function revoke() {
    if (!confirm(`Отозвать подписку у "${team.name}"? Команда уйдёт на free.`)) return;
    setBusy(true);
    try {
      await adminSetSubscription(team.id, { plan: "free", until: "" });
      onSaved();
    } catch (e) {
      onError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-background/70 backdrop-blur-sm flex items-center justify-center p-4 animate-fade-in">
      <div className="relative w-full max-w-2xl rounded-xl border bg-card shadow-2xl max-h-[90vh] overflow-y-auto">
        <button onClick={onClose} className="absolute top-3 right-3 rounded-md p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors">
          <X className="h-4 w-4" />
        </button>
        <div className="p-6 space-y-5">
          <div>
            <span className="eyebrow">Subscription</span>
            <h3 className="display-head text-2xl mt-2">{team.name}</h3>
            <p className="text-xs text-muted-foreground mt-1">{team.owner_email ?? "—"}</p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">План</label>
              <select
                value={plan}
                onChange={(e) => setPlan(e.target.value)}
                className="w-full rounded-md border bg-background px-3 py-2 text-sm font-mono"
              >
                <option value="free">free</option>
                <option value="pro">pro</option>
                <option value="team">team</option>
                <option value="enterprise">enterprise</option>
              </select>
            </div>
            <div className="space-y-1">
              <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">Активна до</label>
              <input
                type="date"
                value={until}
                onChange={(e) => setUntil(e.target.value)}
                className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              />
              <div className="flex gap-1.5 pt-1">
                <button onClick={() => quickExtend(1)} className="text-[11px] font-mono text-muted-foreground hover:text-foreground">+1m</button>
                <button onClick={() => quickExtend(3)} className="text-[11px] font-mono text-muted-foreground hover:text-foreground">+3m</button>
                <button onClick={() => quickExtend(6)} className="text-[11px] font-mono text-muted-foreground hover:text-foreground">+6m</button>
                <button onClick={() => quickExtend(12)} className="text-[11px] font-mono text-muted-foreground hover:text-foreground">+1y</button>
              </div>
            </div>
          </div>

          <div className="space-y-1">
            <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">Заметка (видна owner'у)</label>
            <input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="напр. «Custom deal до конца года»"
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            />
          </div>

          {plan !== "free" && (
            <div className="rounded-lg border bg-muted/20 p-4 space-y-3">
              <label className="flex items-center gap-2 text-sm font-medium">
                <input type="checkbox" checked={recordPayment} onChange={(e) => setRecordPayment(e.target.checked)} />
                Записать платёж
              </label>
              {recordPayment && (
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                  <input
                    type="number"
                    min="0"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    placeholder="сумма (центы)"
                    className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
                  />
                  <input
                    value={currency}
                    onChange={(e) => setCurrency(e.target.value.toUpperCase())}
                    placeholder="USD"
                    maxLength={3}
                    className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono uppercase"
                  />
                  <select
                    value={method}
                    onChange={(e) => setMethod(e.target.value)}
                    className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
                  >
                    <option value="manual_transfer">manual_transfer</option>
                    <option value="cash">cash</option>
                    <option value="stripe">stripe</option>
                    <option value="other">other</option>
                  </select>
                  <input
                    value={paymentNote}
                    onChange={(e) => setPaymentNote(e.target.value)}
                    placeholder="ref / note"
                    className="rounded-md border bg-background px-2 py-1.5 text-sm"
                  />
                </div>
              )}
              <p className="text-[11px] text-muted-foreground font-mono">
                Сумма в центах: 5000 RUB → введи 500000. Запись добавится в журнал платежей.
              </p>
            </div>
          )}

          <div className="flex items-center justify-between pt-2 border-t">
            <Button variant="ghost" onClick={revoke} disabled={busy} className="text-destructive hover:bg-destructive/10">
              Отозвать (вернуть на free)
            </Button>
            <div className="flex gap-2">
              <Button variant="outline" onClick={onClose}>Отмена</Button>
              <Button onClick={save} disabled={busy}>{busy ? "..." : "Сохранить"}</Button>
            </div>
          </div>

          {payments.length > 0 && (
            <div className="pt-4 border-t">
              <span className="eyebrow">История платежей · {payments.length}</span>
              <ul className="mt-3 space-y-1.5">
                {payments.map((p) => (
                  <li key={p.id} className="flex items-baseline justify-between text-xs font-mono">
                    <span className="text-muted-foreground">{formatDate(p.paid_at, tz)}</span>
                    <span>
                      {(p.amount_cents / 100).toFixed(2)} {p.currency}
                      <span className="text-muted-foreground"> · {p.method}</span>
                      {p.note && <span className="text-muted-foreground"> · {p.note}</span>}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
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
