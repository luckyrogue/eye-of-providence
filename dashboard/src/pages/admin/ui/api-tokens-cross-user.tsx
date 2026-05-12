// Cross-user API tokens admin view. Re-uses DataTable pattern.

import { useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  DataTable,
  DataTableColumnHeader,
  DataTableRowActions,
  DropdownMenuItem,
  DropdownMenuLabel,
  EmptyState,
  useConfirm,
  type DataTableColumn,
} from "@eop/ui";
import {
  useAdminCrossTokens,
  useAdminRevokeToken,
  type CrossUserToken,
} from "../../../entities/admin";
import { dtLabels } from "../../../shared/lib/data-table-labels";
import { useMutationToast } from "../../../shared/hooks/use-mutation-toast";
import { formatDate } from "../../../shared/lib/tz";

export function APITokensCrossUser({ tz }: { tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const tokens = useAdminCrossTokens();
  const revoke = useAdminRevokeToken();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const onRevoke = useCallback(
    async (tok: CrossUserToken) => {
      const ok = await confirm({
        title: t("app:admin.tokens_cross.revoke_confirm_title", { prefix: tok.prefix }),
        description: t("app:admin.tokens_cross.revoke_confirm_lead"),
        destructive: true,
        confirmText: t("app:admin.tokens_cross.revoke"),
      });
      if (!ok) return;
      await runToast(revoke.mutateAsync(tok.id), {
        success: t("app:admin.tokens_cross.revoked"),
        error: t("app:admin.tokens_cross.revoke_failed"),
      });
    },
    [confirm, revoke, runToast, t],
  );

  const columns = useMemo<DataTableColumn<CrossUserToken>[]>(
    () => [
      {
        accessorKey: "user_email",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.tokens_cross.col_user")} />
        ),
        cell: ({ row }) => <span className="font-mono text-xs">{row.original.user_email}</span>,
      },
      {
        accessorKey: "prefix",
        header: t("app:admin.tokens_cross.col_prefix"),
        cell: ({ row }) => <span className="font-mono text-xs">{row.original.prefix}…</span>,
      },
      {
        accessorKey: "name",
        header: t("app:admin.tokens_cross.col_name"),
        cell: ({ row }) => row.original.name,
      },
      {
        accessorKey: "scope",
        header: t("app:admin.tokens_cross.col_scope"),
        cell: ({ row }) => <span className="text-xs font-mono">{row.original.scope}</span>,
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
        accessorKey: "last_used_at",
        header: t("app:admin.tokens_cross.col_last_used"),
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {row.original.last_used_at ? formatDate(row.original.last_used_at, tz) : "—"}
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
              <DropdownMenuLabel>{t("app:admin.tokens_cross.actions_label")}</DropdownMenuLabel>
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => void onRevoke(row.original)}
                disabled={revoke.isPending}
              >
                {t("app:admin.tokens_cross.revoke")}
              </DropdownMenuItem>
            </DataTableRowActions>
          </div>
        ),
      },
    ],
    [t, tz, revoke.isPending, onRevoke],
  );

  const list = tokens.data ?? [];

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("app:admin.tokens_cross.title")}</div>
          <div className="card-sub">{t("app:admin.tokens_cross.lead", { count: list.length })}</div>
        </div>
      </div>

      {tokens.isError && (
        <div className="text-[13px] mb-3" style={{ color: "hsl(var(--destructive))" }}>
          {tokens.error?.message ?? t("app:admin.error_lead")}
        </div>
      )}

      {tokens.isPending ? (
        <div className="text-[13px] text-muted-foreground py-3">{t("app:admin.loading")}</div>
      ) : list.length === 0 ? (
        <EmptyState
          eyebrow={t("app:admin.tokens_cross.empty_eyebrow")}
          title={t("app:admin.tokens_cross.empty_title")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={list}
          filterableColumn={{
            id: "user_email",
            placeholder: t("app:admin.tokens_cross.filter_email"),
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
