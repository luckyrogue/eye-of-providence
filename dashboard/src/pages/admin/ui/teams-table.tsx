import { Fragment, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState, PlanBadge, useConfirm } from "@eop/ui";
import { CreditCard, Trash2, UserPlus } from "lucide-react";
import { useAdminDeleteTeam, type AdminTeam, type AdminUser } from "../../../entities/admin";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";
import { IconButton } from "./icon-button";
import { AddMemberRow } from "./add-member-row";
import { SubscriptionModal } from "./subscription-modal";

export function TeamsTable({
  teams,
  users,
  tz,
}: {
  teams: AdminTeam[];
  users: AdminUser[];
  tz: string;
}) {
  const { t } = useTranslation("app");
  const deleteTeam = useAdminDeleteTeam();
  const runToast = useMutationToast();
  const confirm = useConfirm();
  const [adding, setAdding] = useState<string | null>(null);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<"member" | "admin" | "owner">("member");
  const [subTeam, setSubTeam] = useState<AdminTeam | null>(null);

  async function destroy(teamID: string, name: string) {
    const ok = await confirm({
      title: t("admin.team_delete_confirm_title", { name, defaultValue: `Delete "${name}"?` }),
      description: t("admin.team_delete_confirm_lead", {
        defaultValue: "Will erase all members, projects, invites, commit history. Irreversible.",
      }),
      typeToConfirm: name,
      destructive: true,
      confirmText: t("team_detail.settings_danger_confirm"),
    });
    if (!ok) return;
    await runToast(deleteTeam.mutateAsync(teamID), {
      success: t("admin.team_deleted", { defaultValue: "Team deleted" }),
      error: t("admin.team_delete_failed", { defaultValue: "Failed to delete" }),
    });
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
          <CardTitle className="font-display tracking-tight">
            {t("admin.all_teams_title", { defaultValue: "All companies" })}
          </CardTitle>
          <CardDescription>
            {t("admin.all_teams_lead", { count: teams.length, defaultValue: `${teams.length} companies in the system` })}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {teams.length === 0 ? (
            <EmptyState
              eyebrow={t("admin.all_teams_empty_eyebrow", { defaultValue: "No companies" })}
              title={t("admin.all_teams_empty_title", { defaultValue: "No companies yet" })}
            />
          ) : (
            <div className="overflow-x-auto rounded-md border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="py-2.5 px-3 text-left">{t("admin.teams_table_name")}</th>
                    <th className="py-2.5 px-3 text-left">{t("admin.teams_subscribe")}</th>
                    <th className="py-2.5 px-3 text-right">{t("admin.teams_table_members")}</th>
                    <th className="py-2.5 px-3 text-left">{t("admin.teams_table_owner")}</th>
                    <th className="py-2.5 px-3 text-left">{t("admin.teams_table_created")}</th>
                    <th className="py-2.5 px-3 text-right" />
                  </tr>
                </thead>
                <tbody>
                  {teams.map((team) => (
                    <Fragment key={team.id}>
                      <tr className="border-t hover:bg-muted/30">
                        <td className="py-2 px-3 font-medium">{team.name}</td>
                        <td className="py-2 px-3">
                          <PlanBadge plan={team.subscription_plan} until={team.subscription_until} />
                        </td>
                        <td className="py-2 px-3 text-right tabular-nums">{team.member_count}</td>
                        <td className="py-2 px-3 text-xs text-muted-foreground">{team.owner_email ?? "—"}</td>
                        <td className="py-2 px-3 text-xs text-muted-foreground">{formatDate(team.created_at, tz)}</td>
                        <td className="py-2 px-3 text-right">
                          <div className="flex justify-end gap-1">
                            <IconButton
                              title={t("admin.team_subscribe_btn", { defaultValue: "Manage subscription" })}
                              onClick={() => setSubTeam(team)}
                            >
                              <CreditCard className="h-3.5 w-3.5" />
                            </IconButton>
                            <IconButton
                              title={t("admin.team_add_member_btn", { defaultValue: "Add member" })}
                              onClick={() => setAdding(adding === team.id ? null : team.id)}
                            >
                              <UserPlus className="h-3.5 w-3.5" />
                            </IconButton>
                            <IconButton
                              danger
                              title={t("admin.team_delete_btn", { defaultValue: "Delete company" })}
                              onClick={() => destroy(team.id, team.name)}
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </IconButton>
                          </div>
                        </td>
                      </tr>
                      {adding === team.id && (
                        <AddMemberRow
                          teamID={team.id}
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
