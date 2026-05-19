import { useCallback, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  DataTable,
  DataTableColumnHeader,
  DataTableRowActions,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  EmptyState,
  PlanBadge,
  useConfirm,
  type DataTableColumn,
} from "@eop/ui";
import type { AdminTeam, AdminUser } from "@/entities/admin";
import { useAdminDeleteTeam } from "@/entities/admin";
import { AddMemberForm } from "@/features/admin-add-team-member";
import { SubscriptionModal } from "@/features/admin-set-subscription";
import { dtLabels } from "@/shared/lib/data-table-labels";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";
import { formatDate } from "@/shared/lib/tz";
import { TeamDrawer } from "./team-drawer";

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
  const [subTeam, setSubTeam] = useState<AdminTeam | null>(null);
  const [drawerTeam, setDrawerTeam] = useState<AdminTeam | null>(null);
  const confirm = useConfirm();
  const deleteTeam = useAdminDeleteTeam();
  const runToast = useMutationToast();

  const destroy = useCallback(
    async (team: AdminTeam) => {
      const ok = await confirm({
        title: t("app:admin.team_delete_confirm_title", { name: team.name }),
        description: t("app:admin.team_delete_confirm_lead"),
        typeToConfirm: team.name,
        destructive: true,
        confirmText: t("app:team_detail.settings_danger_confirm"),
      });
      if (!ok) return;
      await runToast(deleteTeam.mutateAsync(team.id), {
        success: t("app:admin.team_deleted"),
        error: t("app:admin.team_delete_failed"),
      });
    },
    [confirm, deleteTeam, runToast, t],
  );

  const columns = useMemo<DataTableColumn<AdminTeam>[]>(
    () => [
      {
        accessorKey: "name",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.teams_table_name")} />
        ),
        cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
      },
      {
        id: "plan",
        accessorFn: (row) => row.subscription_plan,
        header: t("app:admin.teams_subscribe"),
        enableSorting: false,
        cell: ({ row }) => (
          <PlanBadge
            plan={row.original.subscription_plan}
            until={row.original.subscription_until}
            untilLabel={t("common:plan_badge.until")}
            expiredLabel={t("common:plan_badge.expired")}
          />
        ),
      },
      {
        accessorKey: "member_count",
        header: ({ column }) => (
          <DataTableColumnHeader
            column={column}
            title={t("app:admin.teams_table_members")}
            align="right"
          />
        ),
        cell: ({ row }) => (
          <div className="text-right tabular-nums">{row.original.member_count}</div>
        ),
      },
      {
        accessorKey: "owner_email",
        header: t("app:admin.teams_table_owner"),
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">{row.original.owner_email ?? "—"}</span>
        ),
      },
      {
        accessorKey: "created_at",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.teams_table_created")} />
        ),
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {formatDate(row.original.created_at, tz)}
          </span>
        ),
      },
      {
        id: "actions",
        enableHiding: false,
        enableSorting: false,
        header: () => null,
        cell: ({ row }) => (
          <div className="flex justify-end">
            <DataTableRowActions triggerLabel={t("common:data_table.open_menu")}>
              <DropdownMenuLabel>{t("app:admin.team_actions_label")}</DropdownMenuLabel>
              <DropdownMenuItem onClick={() => setDrawerTeam(row.original)}>
                {t("app:admin.team_manage_btn")}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setSubTeam(row.original)}>
                {t("app:admin.team_subscribe_btn")}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => row.toggleExpanded()}>
                {row.getIsExpanded()
                  ? t("app:admin.team_add_member_close")
                  : t("app:admin.team_add_member_btn")}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                disabled={deleteTeam.isPending}
                onClick={() => void destroy(row.original)}
              >
                {t("app:admin.team_delete_btn")}
              </DropdownMenuItem>
            </DataTableRowActions>
          </div>
        ),
      },
    ],
    [t, tz, deleteTeam.isPending, destroy],
  );

  return (
    <>
      <div className="eop-card">
        <div className="card-head">
          <div>
            <div className="card-title">{t("app:admin.all_teams_title")}</div>
            <div className="card-sub">{t("app:admin.all_teams_lead", { count: teams.length })}</div>
          </div>
        </div>
        {teams.length === 0 ? (
          <EmptyState
            eyebrow={t("app:admin.all_teams_empty_eyebrow")}
            title={t("app:admin.all_teams_empty_title")}
          />
        ) : (
          <DataTable
            columns={columns}
            data={teams}
            filterableColumn={{
              id: "name",
              placeholder: t("app:admin.teams_filter_name"),
            }}
            enableColumnVisibility
            enablePagination
            pageSize={20}
            labels={dtLabels(t)}
            getRowCanExpand={() => true}
            renderSubComponent={(row) => (
              <AddMemberForm
                teamID={row.original.id}
                users={users}
                onCancel={() => row.toggleExpanded(false)}
                onAdded={() => row.toggleExpanded(false)}
              />
            )}
          />
        )}
      </div>

      <SubscriptionModal team={subTeam} tz={tz} onClose={() => setSubTeam(null)} />
      <TeamDrawer team={drawerTeam} tz={tz} onClose={() => setDrawerTeam(null)} />
    </>
  );
}
