import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  DataTable,
  EmptyState,
  type DataTableColumn,
} from "@eop/ui";
import { useIngestDemo, useRecent, type EventRow } from "../../../entities/event";
import { dtLabels } from "../../../shared/lib/data-table-labels";
import { formatTime } from "../../../shared/lib/tz";
import { CategoryBadge } from "./category-badge";
export function EventsCard({ tz }: { tz: string }) {
  const { t } = useTranslation(["app", "common"]);
  const events = useRecent(20);
  const sendDemo = useIngestDemo();
  const eventsList: EventRow[] = events.data ?? [];
  const columns = useMemo<DataTableColumn<EventRow>[]>(
    () => [
      {
        accessorKey: "ts",
        header: t("app:dashboard.table_time"),
        cell: ({ row }) => (
          <span className="font-mono text-xs">{formatTime(row.original.ts, tz)}</span>
        ),
      },
      {
        accessorKey: "app_bundle",
        header: t("app:dashboard.table_app"),
        cell: ({ row }) => row.original.app_bundle,
      },
      {
        accessorKey: "category",
        header: t("app:dashboard.table_category"),
        cell: ({ row }) => <CategoryBadge cat={row.original.category} />,
      },
      {
        accessorKey: "source",
        header: t("app:dashboard.table_source"),
        cell: ({ row }) => (
          <span className="text-muted-foreground">
            {t(`app:dashboard.source.${row.original.source}` as const)}
          </span>
        ),
      },
      {
        accessorKey: "ai_provider",
        header: t("app:dashboard.table_provider"),
        cell: ({ row }) => (
          <span className="text-muted-foreground">{row.original.ai_provider ?? "—"}</span>
        ),
      },
      {
        accessorKey: "duration_ms",
        header: () => <div className="text-right">{t("app:dashboard.table_duration")}</div>,
        cell: ({ row }) => (
          <div className="text-right tabular-nums">
            {(row.original.duration_ms / 1000).toFixed(1)} {t("app:dashboard.duration_unit_s")}
          </div>
        ),
      },
    ],
    [t, tz],
  );
  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">
          {t("app:dashboard.events_title")}
        </CardTitle>
        <CardDescription>{t("app:dashboard.events_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex gap-2 mb-4">
          <Button onClick={() => sendDemo.mutate()} disabled={sendDemo.isPending} size="sm">
            {t("app:dashboard.events_demo")}
          </Button>
          <Button
            onClick={() => events.refetch()}
            disabled={events.isFetching}
            size="sm"
            variant="outline"
          >
            {t("app:dashboard.events_refresh")}
          </Button>
        </div>
        {eventsList.length === 0 && !events.isPending ? (
          <EmptyState
            eyebrow={t("app:dashboard.events_empty_eyebrow")}
            title={t("app:dashboard.events_empty_title")}
            description={t("app:dashboard.events_empty_lead")}
          />
        ) : (
          <DataTable
            columns={columns}
            data={eventsList}
            isLoading={events.isPending}
            labels={dtLabels(t)}
          />
        )}
      </CardContent>
    </Card>
  );
}
