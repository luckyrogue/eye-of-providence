import { useEffect, useState, Fragment } from "react";
import { useForm } from "react-hook-form";
import {
  Button,
  Card, CardContent, CardDescription, CardHeader, CardTitle,
  EmptyState, Eyebrow, Input, Modal, PlanBadge, Select, StatTile, Tab, TabBar,
  useConfirm,
} from "@eop/ui";
import { Building2, CreditCard, Crown, Trash2, UserPlus, Users } from "lucide-react";
import { formatDate } from "./utils/tz";
import { useMutationToast } from "./hooks/useMutationToast";
import {
  useAdminStats, useAdminTeams, useAdminUsers, useAdminPayments,
  useAdminDeleteTeam, useAdminDeleteUser, useAdminUpdateUser, useAdminAddMember, useAdminSetSubscription,
  type AdminStats, type AdminTeam, type AdminUser,
} from "./api/admin";

type TabKey = "overview" | "teams" | "users";

export function Admin({ tz }: { tz: string }) {
  const [tab, setTab] = useState<TabKey>("overview");
  const stats = useAdminStats();
  const teams = useAdminTeams();
  const users = useAdminUsers();

  return (
    <div className="space-y-4">
      <div className="flex items-baseline justify-between">
        <div>
          <Eyebrow>Super admin</Eyebrow>
          <h2 className="display-head text-3xl mt-2">Управление платформой</h2>
        </div>
        <TabBar>
          <Tab active={tab === "overview"} onClick={() => setTab("overview")}>Обзор</Tab>
          <Tab active={tab === "teams"} onClick={() => setTab("teams")}>Компании</Tab>
          <Tab active={tab === "users"} onClick={() => setTab("users")}>Пользователи</Tab>
        </TabBar>
      </div>

      {tab === "overview" && stats.data && <Overview stats={stats.data} />}
      {tab === "teams" && <TeamsTable teams={teams.data ?? []} users={users.data ?? []} tz={tz} />}
      {tab === "users" && <UsersTable users={users.data ?? []} tz={tz} />}
    </div>
  );
}

function Overview({ stats }: { stats: AdminStats }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <StatTile label="Пользователей" value={stats.users_total} icon={<Users className="h-4 w-4" />} />
      <StatTile
        label="Компаний"
        value={stats.teams_total}
        hint={stats.beta_limit > 0 ? `beta limit · ${stats.beta_limit}` : "лимит снят"}
        icon={<Building2 className="h-4 w-4" />}
      />
      <StatTile label="Membership-связей" value={stats.members_total} icon={<Crown className="h-4 w-4" />} />
    </div>
  );
}

function TeamsTable({ teams, users, tz }: { teams: AdminTeam[]; users: AdminUser[]; tz: string }) {
  const deleteTeam = useAdminDeleteTeam();
  const addMember = useAdminAddMember();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const [adding, setAdding] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<"member" | "admin" | "owner">("member");
  const [subTeam, setSubTeam] = useState<AdminTeam | null>(null);

  async function destroy(teamID: string, name: string) {
    const ok = await confirm({
      title: `Удалить «${name}»?`,
      description: "Будут стерты все участники, проекты, инвайты, история коммитов. Необратимо.",
      typeToConfirm: name,
      destructive: true,
      confirmText: "Удалить навсегда",
    });
    if (!ok) return;
    await runToast(deleteTeam.mutateAsync(teamID), { success: "Компания удалена", error: "Не удалось удалить" });
  }

  async function onAdd(teamID: string) {
    if (!email.trim()) return;
    const ok = await runToast(
      addMember.mutateAsync({ teamID, email: email.trim(), role }),
      { success: "Участник добавлен", error: "Не удалось добавить" },
    );
    if (ok !== null) {
      setAdding(null);
      setEmail("");
      setRole("member");
    }
  }

  return (
    <>
      <Card className="card-hover">
        <CardHeader>
          <CardTitle className="font-display tracking-tight">Все компании</CardTitle>
          <CardDescription>{teams.length} компаний в системе</CardDescription>
        </CardHeader>
        <CardContent>
          {teams.length === 0 ? (
            <EmptyState eyebrow="No companies" title="Пока нет ни одной компании" />
          ) : (
            <div className="overflow-x-auto rounded-md border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="py-2.5 px-3 text-left">Название</th>
                    <th className="py-2.5 px-3 text-left">Подписка</th>
                    <th className="py-2.5 px-3 text-right">Members</th>
                    <th className="py-2.5 px-3 text-left">Owner</th>
                    <th className="py-2.5 px-3 text-left">Создана</th>
                    <th className="py-2.5 px-3 text-right" />
                  </tr>
                </thead>
                <tbody>
                  {teams.map((t) => (
                    <Fragment key={t.id}>
                      <tr className="border-t hover:bg-muted/30">
                        <td className="py-2 px-3 font-medium">{t.name}</td>
                        <td className="py-2 px-3">
                          <PlanBadge plan={t.subscription_plan} until={t.subscription_until} />
                        </td>
                        <td className="py-2 px-3 text-right tabular-nums">{t.member_count}</td>
                        <td className="py-2 px-3 text-xs text-muted-foreground">{t.owner_email ?? "—"}</td>
                        <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(t.created_at, tz)}</td>
                        <td className="py-2 px-3 text-right">
                          <div className="flex justify-end gap-1">
                            <IconButton title="Управление подпиской" onClick={() => setSubTeam(t)}>
                              <CreditCard className="h-3.5 w-3.5" />
                            </IconButton>
                            <IconButton title="Добавить участника" onClick={() => setAdding(adding === t.id ? null : t.id)}>
                              <UserPlus className="h-3.5 w-3.5" />
                            </IconButton>
                            <IconButton danger title="Удалить компанию" onClick={() => destroy(t.id, t.name)}>
                              <Trash2 className="h-3.5 w-3.5" />
                            </IconButton>
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
                              <Button size="sm" onClick={() => onAdd(t.id)} disabled={addMember.isPending || !email.trim()}>
                                Добавить
                              </Button>
                              <Button size="sm" variant="ghost" onClick={() => { setAdding(null); setEmail(""); }}>
                                Отмена
                              </Button>
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <SubscriptionModal team={subTeam} tz={tz} onClose={() => setSubTeam(null)} />
    </>
  );
}

function UsersTable({ users, tz }: { users: AdminUser[]; tz: string }) {
  const update = useAdminUpdateUser();
  const del = useAdminDeleteUser();
  const runToast = useMutationToast();
  const confirm = useConfirm();

  async function setRole(uid: string, role: string) {
    await runToast(update.mutateAsync({ userID: uid, payload: { global_role: role } }), {
      success: "Роль обновлена",
      error: "Не удалось изменить роль",
    });
  }

  async function destroy(uid: string, email: string) {
    const ok = await confirm({
      title: `Удалить ${email}?`,
      description: "Все события и отчёты будут стерты. Необратимо.",
      destructive: true,
      confirmText: "Удалить",
    });
    if (!ok) return;
    await runToast(del.mutateAsync(uid), { success: "Пользователь удалён", error: "Не удалось удалить" });
  }

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">Все пользователи</CardTitle>
        <CardDescription>{users.length} учётных записей</CardDescription>
      </CardHeader>
      <CardContent>
        {users.length === 0 ? (
          <EmptyState eyebrow="No users" title="Пока никого нет" />
        ) : (
          <div className="overflow-x-auto rounded-md border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2.5 px-3 text-left">Email</th>
                  <th className="py-2.5 px-3 text-left">Имя</th>
                  <th className="py-2.5 px-3 text-left">Global role</th>
                  <th className="py-2.5 px-3 text-right">Команды</th>
                  <th className="py-2.5 px-3 text-left">Создан</th>
                  <th className="py-2.5 px-3 text-right" />
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id} className="border-t hover:bg-muted/30">
                    <td className="py-2 px-3 font-mono text-xs">{u.email}</td>
                    <td className="py-2 px-3">{u.display_name}</td>
                    <td className="py-2 px-3">
                      <Select
                        mono
                        value={u.global_role}
                        onChange={(e) => setRole(u.id, e.target.value)}
                        disabled={update.isPending}
                        className="px-2 py-0.5 text-xs"
                      >
                        <option value="user">user</option>
                        <option value="super_admin">super_admin</option>
                      </Select>
                    </td>
                    <td className="py-2 px-3 text-right tabular-nums">{u.teams_count ?? 0}</td>
                    <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(u.created_at, tz)}</td>
                    <td className="py-2 px-3 text-right">
                      <IconButton danger title="Удалить пользователя" onClick={() => destroy(u.id, u.email)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </IconButton>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function IconButton({
  onClick,
  title,
  children,
  danger,
}: {
  onClick?: () => void;
  title?: string;
  children: React.ReactNode;
  danger?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      className={`rounded-md p-1.5 text-muted-foreground transition-colors ${
        danger
          ? "hover:text-destructive hover:bg-destructive/10"
          : "hover:text-foreground hover:bg-secondary"
      }`}
    >
      {children}
    </button>
  );
}

// --- Subscription modal ---

interface SubscriptionForm {
  plan: "free" | "pro" | "team" | "enterprise";
  until: string;
  note: string;
  recordPayment: boolean;
  amount: string;
  currency: string;
  method: string;
  paymentNote: string;
}

function SubscriptionModal({ team, tz, onClose }: { team: AdminTeam | null; tz: string; onClose: () => void }) {
  const setSub = useAdminSetSubscription();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const payments = useAdminPayments(team?.id ?? null);

  const { register, handleSubmit, watch, setValue, reset } = useForm<SubscriptionForm>({
    defaultValues: {
      plan: "free",
      until: "",
      note: "",
      recordPayment: true,
      amount: "",
      currency: "USD",
      method: "manual_transfer",
      paymentNote: "",
    },
  });

  // При смене team — заполнить форму актуальными значениями.
  useEffect(() => {
    if (team) {
      reset({
        plan: (team.subscription_plan as SubscriptionForm["plan"]) || "free",
        until: team.subscription_until ? team.subscription_until.slice(0, 10) : "",
        note: team.subscription_note ?? "",
        recordPayment: true,
        amount: "",
        currency: "USD",
        method: "manual_transfer",
        paymentNote: "",
      });
    }
  }, [team, reset]);

  function quickExtend(months: number) {
    const current = watch("until");
    const base = current ? new Date(current) : new Date();
    if (!current || base < new Date()) base.setTime(Date.now());
    base.setMonth(base.getMonth() + months);
    setValue("until", base.toISOString().slice(0, 10), { shouldDirty: true });
  }

  async function onSave(values: SubscriptionForm) {
    if (!team) return;
    const payload: Parameters<typeof setSub.mutateAsync>[0]["payload"] = {
      plan: values.plan,
      until: values.until ? new Date(values.until + "T23:59:59Z").toISOString() : "",
      note: values.note,
    };
    const amt = parseInt(values.amount, 10);
    if (values.recordPayment && values.plan !== "free" && amt > 0 && values.until) {
      payload.payment = {
        amount_cents: amt,
        currency: values.currency,
        method: values.method,
        note: values.paymentNote,
        covers_until: new Date(values.until + "T23:59:59Z").toISOString(),
      };
    }
    const ok = await runToast(setSub.mutateAsync({ teamID: team.id, payload }), {
      success: "Подписка обновлена",
      error: "Не удалось обновить подписку",
    });
    if (ok !== null) onClose();
  }

  async function revoke() {
    if (!team) return;
    const proceed = await confirm({
      title: `Отозвать подписку у «${team.name}»?`,
      description: "Команда уйдёт на free-тариф. Это можно обратить, выдав подписку заново.",
      destructive: true,
      confirmText: "Отозвать",
    });
    if (!proceed) return;
    const ok = await runToast(
      setSub.mutateAsync({ teamID: team.id, payload: { plan: "free", until: "" } }),
      { success: "Подписка отозвана", error: "Не удалось отозвать" },
    );
    if (ok !== null) onClose();
  }

  if (!team) return null;
  const plan = watch("plan");
  const recordPayment = watch("recordPayment");

  return (
    <Modal open={!!team} onClose={onClose}>
      <form onSubmit={handleSubmit(onSave)} className="p-6 space-y-5">
        <div>
          <Eyebrow>Subscription</Eyebrow>
          <h3 className="display-head text-2xl mt-2">{team.name}</h3>
          <p className="text-xs text-muted-foreground mt-1">{team.owner_email ?? "—"}</p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Select label="План" mono {...register("plan")} className="w-full px-3 py-2">
            <option value="free">free</option>
            <option value="pro">pro</option>
            <option value="team">team</option>
            <option value="enterprise">enterprise</option>
          </Select>
          <div className="space-y-1">
            <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">Активна до</label>
            <input
              type="date"
              {...register("until")}
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            />
            <div className="flex gap-1.5 pt-1">
              {[1, 3, 6, 12].map((m) => (
                <button
                  key={m}
                  type="button"
                  onClick={() => quickExtend(m)}
                  className="text-[11px] font-mono text-muted-foreground hover:text-foreground"
                >
                  +{m === 12 ? "1y" : `${m}m`}
                </button>
              ))}
            </div>
          </div>
        </div>

        <Input
          label="Заметка (видна owner'у)"
          placeholder="напр. «Custom deal до конца года»"
          {...register("note")}
        />

        {plan !== "free" && (
          <div className="rounded-lg border bg-muted/20 p-4 space-y-3">
            <label className="flex items-center gap-2 text-sm font-medium">
              <input type="checkbox" {...register("recordPayment")} />
              Записать платёж
            </label>
            {recordPayment && (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                <input
                  type="number"
                  min="0"
                  {...register("amount")}
                  placeholder="сумма (центы)"
                  className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
                />
                <input
                  {...register("currency")}
                  placeholder="USD"
                  maxLength={3}
                  className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono uppercase"
                />
                <select
                  {...register("method")}
                  className="rounded-md border bg-background px-2 py-1.5 text-sm font-mono"
                >
                  <option value="manual_transfer">manual_transfer</option>
                  <option value="cash">cash</option>
                  <option value="stripe">stripe</option>
                  <option value="other">other</option>
                </select>
                <input
                  {...register("paymentNote")}
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
          <Button type="button" variant="ghost" onClick={revoke} disabled={setSub.isPending} className="text-destructive hover:bg-destructive/10">
            Отозвать (вернуть на free)
          </Button>
          <div className="flex gap-2">
            <Button type="button" variant="outline" onClick={onClose}>Отмена</Button>
            <Button type="submit" disabled={setSub.isPending}>{setSub.isPending ? "..." : "Сохранить"}</Button>
          </div>
        </div>

        {payments.data && payments.data.length > 0 && (
          <div className="pt-4 border-t">
            <Eyebrow>История платежей · {payments.data.length}</Eyebrow>
            <ul className="mt-3 space-y-1.5">
              {payments.data.map((p) => (
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
      </form>
    </Modal>
  );
}

