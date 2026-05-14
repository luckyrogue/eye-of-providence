import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { DataTable, type DataTableColumn } from "@eop/ui";
import type { Device } from "../../../entities/device";
import { RevokeDeviceButton } from "../../../features/device-revoke";
import { DeviceKindBadge } from "./device-kind-badge";
export function DevicesTable({ devices, isLoading }: { devices: Device[]; isLoading: boolean }) {
  const { t } = useTranslation("developer");
  const columns = useMemo<DataTableColumn<Device>[]>(
    () => [
      {
        id: "kind",
        header: () => <span className="text-xs uppercase tracking-wider">Kind</span>,
        cell: ({ row }) => <DeviceKindBadge kind={row.original.kind} />,
        size: 180,
      },
      {
        id: "name",
        header: () => <span className="text-xs uppercase tracking-wider">Name</span>,
        cell: ({ row }) => (
          <div className="min-w-0">
            <div className="font-medium text-sm truncate">{row.original.name}</div>
            <code className="font-mono text-[11px] text-muted-foreground">
              {row.original.prefix}…
            </code>
          </div>
        ),
      },
      {
        id: "last_used",
        header: () => <span className="text-xs uppercase tracking-wider">Last used</span>,
        cell: ({ row }) => (
          <span className="text-xs text-muted-foreground">
            {row.original.last_used_at
              ? t("devices_last_used", { at: new Date(row.original.last_used_at).toLocaleString() })
              : t("devices_never_used")}
          </span>
        ),
        size: 220,
      },
      {
        id: "actions",
        header: () => null,
        cell: ({ row }) => (
          <div className="flex justify-end">
            <RevokeDeviceButton device={row.original} />
          </div>
        ),
        size: 120,
      },
    ],
    [t],
  );
  return (
    <DataTable
      columns={columns}
      data={devices}
      isLoading={isLoading}
      emptyState={<p className="text-sm text-muted-foreground py-4 px-2">{t("devices_empty")}</p>}
    />
  );
}
