import { Fragment, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState, IconButton, PlanBadge } from "@eop/ui";
import { CreditCard, UserPlus } from "lucide-react";
import type { AdminTeam, AdminUser } from "../../../entities/admin";
import { AddMemberRow } from "../../../features/admin-add-team-member";
import { DeleteTeamButton } from "../../../features/admin-delete-team";
import { SubscriptionModal } from "../../../features/admin-set-subscription";
import { formatDate } from "../../../shared/lib/tz";

export function TeamsTable({
  teams,
  users,
  tz,
}: {
  teams: AdminTeam[];
  users: AdminUser[];
  tz: string;
}) {
  const { t } = useTranslation(["app", "common"]);
  const [adding, setAdding] = useState<string | null>(null);
  const [subTeam, setSubTeam] = useState<AdminTeam | null>(null);

  return (
    <>
      <Card className="card-hover">
        <CardHeader>
          <CardTitle className="font-display tracking-tight">
            {t("admin.all_teams_title", { defaultValue: "All companies" })}
          </CardTitle>
          <CardDescription>
            {t("admin.all_teams_lead", {
              count: teams.length,
              defaultValue: `${teams.length} companies in the system`,
            })}
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
                          <PlanBadge
                            plan={team.subscription_plan}
                            until={team.subscription_until}
                            untilLabel={t("common:plan_badge.until", { defaultValue: "until" })}
                            expiredLabel={t("common:plan_badge.expired", { defaultValue: "expired" })}
                          />
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
                            <DeleteTeamButton teamID={team.id} name={team.name} />
                          </div>
                        </td>
                      </tr>
                      {adding === team.id && (
                        <AddMemberRow
                          teamID={team.id}
                          users={users}
                          onCancel={() => setAdding(null)}
                          onAdded={() => setAdding(null)}
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
