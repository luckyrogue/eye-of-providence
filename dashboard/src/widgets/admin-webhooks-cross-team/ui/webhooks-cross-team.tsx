// Cross-team webhooks admin view. Re-uses DataTable pattern from users-table.

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
  useAdminCrossWebhooks,
  useAdminRevokeWebhook,
  type CrossTeamWebhook,
} from "@/entities/admin";
import { dtLabels } from "@/shared/lib/data-table-labels";
import { useMutationToast } from "@/shared/hooks/use-mutation-toast";
import { formatDate } from "@/shared/lib/tz";

export function WebhooksCrossTeam({ tz }: { tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const webhooks = useAdminCrossWebhooks();
  const revoke = useAdminRevokeWebhook();
  const confirm = useConfirm();
  const runToast = useMutationToast();

  const onRevoke = useCallback(
    async (wh: CrossTeamWebhook) => {
      const ok = await confirm({
        title: t("app:admin.webhooks_cross.revoke_confirm_title", { url: wh.url }),
        description: t("app:admin.webhooks_cross.revoke_confirm_lead"),
        destructive: true,
        confirmText: t("app:admin.webhooks_cross.revoke"),
      });
      if (!ok) return;
      await runToast(revoke.mutateAsync({ teamID: wh.team_id, webhookID: wh.id }), {
        success: t("app:admin.webhooks_cross.revoked"),
        error: t("app:admin.webhooks_cross.revoke_failed"),
      });
    },
    [confirm, revoke, runToast, t],
  );

  const columns = useMemo<DataTableColumn<CrossTeamWebhook>[]>(
    () => [
      {
        accessorKey: "team_name",
        header: ({ column }) => (
          <DataTableColumnHeader column={column} title={t("app:admin.webhooks_cross.col_team")} />
        ),
        cell: ({ row }) => <span className="font-medium">{row.original.team_name}</span>,
      },
      {
        accessorKey: "url",
        header: t("app:admin.webhooks_cross.col_url"),
        cell: ({ row }) => <span className="font-mono text-xs break-all">{row.original.url}</span>,
      },
      {
        accessorKey: "events",
        header: t("app:admin.webhooks_cross.col_events"),
        cell: ({ row }) => (
          <span className="text-xs">{(row.original.events ?? []).join(", ")}</span>
        ),
      },
      {
        accessorKey: "format",
        header: t("app:admin.webhooks_cross.col_format"),
        cell: ({ row }) => (
          <span className="text-xs font-mono uppercase">{row.original.format}</span>
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
              <DropdownMenuLabel>{t("app:admin.webhooks_cross.actions_label")}</DropdownMenuLabel>
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => void onRevoke(row.original)}
                disabled={revoke.isPending}
              >
                {t("app:admin.webhooks_cross.revoke")}
              </DropdownMenuItem>
            </DataTableRowActions>
          </div>
        ),
      },
    ],
    [t, tz, revoke.isPending, onRevoke],
  );

  const list = webhooks.data ?? [];

  return (
    <div className="eop-card">
      <div className="card-head">
        <div>
          <div className="card-title">{t("app:admin.webhooks_cross.title")}</div>
          <div className="card-sub">
            {t("app:admin.webhooks_cross.lead", { count: list.length })}
          </div>
        </div>
      </div>

      {webhooks.isError && (
        <div className="text-[13px] mb-3" style={{ color: "hsl(var(--destructive))" }}>
          {webhooks.error?.message ?? t("app:admin.error_lead")}
        </div>
      )}

      {webhooks.isPending ? (
        <div className="text-[13px] text-muted-foreground py-3">{t("app:admin.loading")}</div>
      ) : list.length === 0 ? (
        <EmptyState
          eyebrow={t("app:admin.webhooks_cross.empty_eyebrow")}
          title={t("app:admin.webhooks_cross.empty_title")}
        />
      ) : (
        <DataTable
          columns={columns}
          data={list}
          filterableColumn={{
            id: "url",
            placeholder: t("app:admin.webhooks_cross.filter_url"),
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
