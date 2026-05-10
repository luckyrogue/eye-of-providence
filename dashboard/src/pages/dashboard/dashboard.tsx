import { useHeatmap, useLanguages, useRecent, useSummary, useTrend } from "../../entities/event";
import { useReports } from "../../entities/report";
import { InstallPWA } from "../../features/install-pwa";
import { InsightsWidget } from "../../widgets/insights";
import { EventsCard } from "./ui/events-card";
import { HeatmapCard } from "./ui/heatmap-card";
import { LanguagesCard } from "./ui/languages-card";
import { ReportsCard } from "./ui/reports-card";
import { StatsRow } from "./ui/stats-row";
import { TrendCard } from "./ui/trend-card";

export function Dashboard({ tz }: { tz: string }) {
  const events = useRecent(20);
  const summary = useSummary(7);
  const languages = useLanguages(30);
  const heatmap = useHeatmap(30, tz);
  const trend = useTrend(30, tz);
  const reports = useReports();

  const sumMap = summary.data ?? {};
  const totalMs = Object.values(sumMap).reduce((a, b) => a + b, 0);
  const aiMs = sumMap["ai"] ?? 0;
  const aiRatio = totalMs ? Math.round((aiMs / totalMs) * 100) : 0;
  const eventsList = events.data ?? [];
  const reportsList = reports.data ?? [];

  return (
    <>
      <StatsRow
        aiRatio={aiRatio}
        totalMinutes={Math.round(totalMs / 60000)}
        eventsCount={eventsList.length}
        reportsCount={reportsList.length}
      />

      <InstallPWA />
      <InsightsWidget />

      <TrendCard points={trend.data ?? []} />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <HeatmapCard cells={heatmap.data ?? []} tz={tz} />
        <LanguagesCard cells={languages.data ?? []} />
      </div>

      <ReportsCard reports={reportsList} tz={tz} />
      <EventsCard tz={tz} />
    </>
  );
}
