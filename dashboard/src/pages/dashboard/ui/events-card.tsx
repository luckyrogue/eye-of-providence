import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, EmptyState } from "@eop/ui";
import { useIngestDemo, useRecent, type EventRow } from "../../../entities/event";
import { formatTime } from "../../../shared/lib/tz";
import { CategoryBadge } from "./category-badge";

export function EventsCard({ tz }: { tz: string }) {
  const { t } = useTranslation("app");
  const events = useRecent(20);
  const sendDemo = useIngestDemo();
  const eventsList: EventRow[] = events.data ?? [];

  return (
    <Card className="card-hover">
      <CardHeader>
        <CardTitle className="font-display tracking-tight">{t("dashboard.events_title")}</CardTitle>
        <CardDescription>{t("dashboard.events_lead")}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex gap-2 mb-4">
          <Button onClick={() => sendDemo.mutate()} disabled={sendDemo.isPending} size="sm">
            {t("dashboard.events_demo")}
          </Button>
          <Button onClick={() => events.refetch()} disabled={events.isFetching} size="sm" variant="outline">
            {t("dashboard.events_refresh")}
          </Button>
        </div>
        {eventsList.length === 0 ? (
          <EmptyState
            eyebrow={t("dashboard.events_empty_eyebrow")}
            title={t("dashboard.events_empty_title")}
            description={t("dashboard.events_empty_lead")}
          />
        ) : (
          <div className="overflow-x-auto rounded-md border">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
                <tr>
                  <th className="py-2.5 px-3">{t("dashboard.table_time")}</th>
                  <th className="py-2.5 px-3">{t("dashboard.table_app")}</th>
                  <th className="py-2.5 px-3">{t("dashboard.table_category")}</th>
                  <th className="py-2.5 px-3">{t("dashboard.table_source")}</th>
                  <th className="py-2.5 px-3">{t("dashboard.table_provider")}</th>
                  <th className="py-2.5 px-3 text-right">{t("dashboard.table_duration")}</th>
                </tr>
              </thead>
              <tbody>
                {eventsList.map((e, i) => (
                  <tr key={i} className="border-t hover:bg-muted/30 transition-colors">
                    <td className="py-2 px-3 font-mono text-xs">{formatTime(e.ts, tz)}</td>
                    <td className="py-2 px-3">{e.app_bundle}</td>
                    <td className="py-2 px-3"><CategoryBadge cat={e.category} /></td>
                    <td className="py-2 px-3 text-muted-foreground">
                      {t(`dashboard.source.${e.source}` as const, { defaultValue: e.source })}
                    </td>
                    <td className="py-2 px-3 text-muted-foreground">{e.ai_provider ?? "—"}</td>
                    <td className="py-2 px-3 text-right tabular-nums">{(e.duration_ms / 1000).toFixed(1)} с</td>
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
