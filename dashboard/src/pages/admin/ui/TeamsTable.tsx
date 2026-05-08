import { Fragment, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState, PlanBadge, useConfirm } from "@eop/ui";
import { CreditCard, Trash2, UserPlus } from "lucide-react";
import { useAdminDeleteTeam, type AdminTeam, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/useMutationToast";
import { formatDate } from "../../../shared/lib/tz";
import { IconButton } from "./IconButton";
import { AddMemberRow } from "./AddMemberRow";
import { SubscriptionModal } from "./SubscriptionModal";

export function TeamsTable({
  teams,
  users,
  tz,
}: {
  teams: AdminTeam[];
  users: AdminUser[];
  tz: string;
}) {
  const deleteTeam = useAdminDeleteTeam();
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

  function resetAdd() {
    setAdding(null);
    setEmail("");
    setRole("member");
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
                        <AddMemberRow
                          teamID={t.id}
                          users={users}
                          email={email}
                          setEmail={setEmail}
                          role={role}
                          setRole={setRole}
                          onCancel={resetAdd}
                          onAdded={resetAdd}
                        />
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
