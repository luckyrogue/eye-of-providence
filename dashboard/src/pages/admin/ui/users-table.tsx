import { useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  DataTable,
  DataTableColumnHeader,
  DataTableRowActions,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  EmptyState,
  type DataTableColumn,
} from "@eop/ui";
import type { AdminUser } from "../../../entities/admin";
import { useAdminDeleteUser } from "../../../entities/admin";
import { useConfirm } from "@eop/ui";
import { UserRoleSelect } from "../../../features/admin-update-user-role";
import { dtLabels } from "../../../shared/lib/data-table-labels";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";
export function UsersTable({ users, tz }: { users: AdminUser[]; tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const confirm = useConfirm();
  const del = useAdminDeleteUser();
  const runToast = useMutationToast();
  const destroy = useCallback(
    async (user: AdminUser) => {
      const ok = await confirm({
        title: t("app:admin.user_delete_confirm_title", { email: user.email }),
        description: t("app:admin.user_delete_confirm_lead"),
        destructive: true,
        confirmText: t("app:admin.users_delete"),
      });
      if (!ok) return;
      await runToast(del.mutateAsync(user.id), {
        success: t("app:admin.users_deleted"),
        error: t("app:admin.users_delete_failed"),
      });
    },
    [confirm, del, runToast, t],
  );
  const columns = useMemo<DataTableColumn<AdminUser>[]>(
    () => [
      {
        accessorKey: "email",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.users_email")} />
        ),
        cell: ({ row }) => <span className="font-mono text-xs">{row.original.email}</span>,
      },
      {
        accessorKey: "display_name",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.users_name")} />
        ),
        cell: ({ row }) => row.original.display_name,
      },
      {
        accessorKey: "global_role",
        header: t("app:admin.users_global_role"),
        enableSorting: false,
        cell: ({ row }) => (
          <UserRoleSelect userID={row.original.id} value={row.original.global_role} />
        ),
      },
      {
        accessorKey: "teams_count",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.users_team")} align="right" />
        ),
        cell: ({ row }) => (
          <div className="text-right tabular-nums">{row.original.teams_count ?? 0}</div>
        ),
      },
      {
        accessorKey: "created_at",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.table_created")} />
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
              <DropdownMenuLabel>{t("app:admin.user_actions_label")}</DropdownMenuLabel>
              <DropdownMenuItem onClick={() => navigator.clipboard.writeText(row.original.email)}>
                {t("app:admin.user_action_copy_email")}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => void destroy(row.original)}
                disabled={del.isPending}
              >
                {t("app:admin.users_delete")}
              </DropdownMenuItem>
            </DataTableRowActions>
          </div>
        ),
      },
    ],
    [t, tz, del.isPending, destroy],
  );
  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("app:admin.all_users_title")}</div>
          <div className="card-sub">{t("app:admin.all_users_lead", { count: users.length })}</div>
        </div>
      </div>
      {users.length === 0 ? (
        <EmptyState
          eyebrow={t("app:admin.all_users_empty_eyebrow")}
          title={t("app:admin.all_users_empty_title")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={users}
          filterableColumn={{
            id: "email",
            placeholder: t("app:admin.users_filter_email"),
          }}
          enableColumnVisibility
          enablePagination
          pageSize={20}
          labels={dtLabels(t)}
        />
      )}
    </div>
  );
}
