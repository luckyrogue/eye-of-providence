import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Badge, Card, CardContent, CardDescription, CardHeader, CardTitle, Button, EmptyState, Eyebrow, StatTile } from "@eop/ui";
import { Activity, Brain, FileText, Sparkles } from "lucide-react";
import { Markdown } from "../../shared/lib/markdown";
import { Heatmap } from "../../shared/ui/charts/heatmap";
import { Languages } from "../../shared/ui/charts/languages";
import { Trend } from "../../shared/ui/charts/trend";
import { formatDate, formatTime, getTz } from "../../shared/lib/tz";
import { useRecent, useSummary, useLanguages, useHeatmap, useTrend, useIngestDemo } from "../../entities/event";
import { useReports, useGenerateReport, type Report } from "../../entities/report";
import { InsightsWidget } from "../../widgets/insights";

export function DashboardRoute() {
  const { t } = useTranslation("app");
  const tz = getTz();
  const events = useRecent(20);
  const summary = useSummary(7);
  const languages = useLanguages(30);
  const heatmap = useHeatmap(30, tz);
  const trend = useTrend(30, tz);
  const reports = useReports();

  const [activeReport, setActiveReport] = useState<Report | null>(null);
  const genReport = useGenerateReport();
  const sendDemo = useIngestDemo();

  // Если отчёт ещё не выбран и есть свежий — показать первый.
  const reportsList = reports.data ?? [];
  const current = activeReport ?? reportsList[0] ?? null;

  const sumMap = summary.data ?? {};
  const totalMs = Object.values(sumMap).reduce((a, b) => a + b, 0);
  const aiMs = sumMap["ai"] ?? 0;
  const aiRatio = totalMs ? Math.round((aiMs / totalMs) * 100) : 0;
  const eventsList = events.data ?? [];

  return (
    <>
      <div className="reveal">
        <div className="flex items-baseline justify-between mb-3">
          <Eyebrow>{t("dashboard.overview")}</Eyebrow>
          <span className="font-mono text-[11px] text-muted-foreground">{t("dashboard.last_7_days")}</span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatTile label={t("dashboard.stat_ai_share")} value={`${aiRatio}%`} hint={t("dashboard.stat_ai_hint")} icon={<Brain className="h-4 w-4 text-purple-500" />} accent="purple" className="reveal reveal-delay-1" />
          <StatTile label={t("dashboard.stat_active")} value={Math.round(totalMs / 60000)} unit={t("dashboard.stat_minutes")} hint={t("dashboard.stat_active_hint_events", { count: eventsList.length })} icon={<Activity className="h-4 w-4 text-blue-500" />} accent="blue" className="reveal reveal-delay-2" />
          <StatTile label={t("dashboard.stat_reports")} value={reportsList.length} hint={t("dashboard.stat_reports_hint")} icon={<FileText className="h-4 w-4 text-amber-500" />} accent="amber" className="reveal reveal-delay-3" />
        </div>
      </div>

      <InsightsWidget />

      <Card className="card-hover">
        <CardHeader>
          <CardTitle className="font-display tracking-tight">{t("dashboard.trend_title")}</CardTitle>
          <CardDescription>{t("dashboard.trend_lead")}</CardDescription>
        </CardHeader>
        <CardContent>
          <Trend points={trend.data ?? []} />
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="font-display tracking-tight">{t("dashboard.heatmap_title")}</CardTitle>
            <CardDescription>{t("dashboard.heatmap_lead", { tz })}</CardDescription>
          </CardHeader>
          <CardContent>
            <Heatmap cells={heatmap.data ?? []} />
          </CardContent>
        </Card>
        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="font-display tracking-tight">{t("dashboard.languages_title")}</CardTitle>
            <CardDescription>{t("dashboard.languages_lead")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Languages cells={languages.data ?? []} top={6} />
          </CardContent>
        </Card>
      </div>

      <Card className="card-hover">
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 font-display tracking-tight">
              <Sparkles className="h-4 w-4 text-purple-500" />
              {t("dashboard.report_title")}
            </CardTitle>
            <CardDescription>{t("dashboard.report_lead")}</CardDescription>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={() => genReport.mutate("weekly")} disabled={genReport.isPending}>
              {genReport.isPending ? "..." : t("dashboard.report_create_weekly")}
            </Button>
            <Button size="sm" variant="outline" onClick={() => genReport.mutate("monthly")} disabled={genReport.isPending}>
              {t("dashboard.report_create_monthly")}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {reportsList.length > 1 && (
            <div className="flex gap-2 flex-wrap">
              {reportsList.map((r) => (
                <button
                  key={r.id}
                  onClick={() => setActiveReport(r)}
                  className={`rounded-md px-3 py-1 text-xs transition-colors ${
                    current?.id === r.id ? "bg-primary text-primary-foreground" : "bg-secondary hover:bg-secondary/80"
                  }`}
                >
                  {r.period}
                </button>
              ))}
            </div>
          )}
          {current ? (
            <div className="rounded-lg border bg-card/50 p-5">
              <Markdown source={current.body_md} />
              <div className="mt-4 pt-4 border-t text-xs text-muted-foreground">
                <span className="font-mono">{current.model}</span> · {formatDate(current.generated_at, tz)}
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t("dashboard.report_empty")}</p>
          )}
        </CardContent>
      </Card>

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
                      <td className="py-2 px-3 text-muted-foreground">{t(`dashboard.source.${e.source}` as const, { defaultValue: e.source })}</td>
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
    </>
  );
}

const CATEGORY_TONES: Record<string, "blue" | "purple" | "amber" | "neutral"> = {
  manual: "blue",
  ai: "purple",
  refactor: "amber",
  idle: "neutral",
  other: "neutral",
  reading: "neutral",
};

function CategoryBadge({ cat }: { cat: string }) {
  const { t } = useTranslation("app");
  return (
    <Badge tone={CATEGORY_TONES[cat] ?? "neutral"}>
      {t(`dashboard.category.${cat}` as const, { defaultValue: cat })}
    </Badge>
  );
}
