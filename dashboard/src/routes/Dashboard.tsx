import { useState } from "react";
import { Badge, Card, CardContent, CardDescription, CardHeader, CardTitle, Button, EmptyState, Eyebrow, SkeletonTable, StatTile } from "@eop/ui";
import { Activity, Brain, FileText, Sparkles } from "lucide-react";
import { Markdown } from "../Markdown";
import { Heatmap } from "../Heatmap";
import { Languages } from "../Languages";
import { Trend } from "../Trend";
import { formatDate, formatTime, getTz } from "../tz";
import {
  useRecent,
  useSummary,
  useLanguages,
  useHeatmap,
  useTrend,
  useReports,
  useGenerateReport,
  useIngestDemo,
} from "../hooks/queries";
import type { Report } from "../api";

const CATEGORY_LABELS: Record<string, string> = {
  manual: "вручную", ai: "AI", refactor: "рефакторинг", idle: "простой", reading: "чтение", other: "прочее",
};
const SOURCE_LABELS: Record<string, string> = { os: "ОС", browser: "браузер", ide: "IDE", cli: "CLI" };

export function DashboardRoute() {
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
          <Eyebrow>Overview</Eyebrow>
          <span className="font-mono text-[11px] text-muted-foreground">last 7 days</span>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatTile label="Доля AI" value={`${aiRatio}%`} hint="за 7 дней" icon={<Brain className="h-4 w-4 text-purple-500" />} accent="purple" className="reveal reveal-delay-1" />
          <StatTile label="Активное время" value={Math.round(totalMs / 60000)} unit="мин" hint={`${eventsList.length} событий за период`} icon={<Activity className="h-4 w-4 text-blue-500" />} accent="blue" className="reveal reveal-delay-2" />
          <StatTile label="Отчёты" value={reportsList.length} hint="сгенерировано" icon={<FileText className="h-4 w-4 text-amber-500" />} accent="amber" className="reveal reveal-delay-3" />
        </div>
      </div>

      <Card className="card-hover">
        <CardHeader>
          <CardTitle className="font-display tracking-tight">Динамика за 30 дней</CardTitle>
          <CardDescription>Время вручную vs с AI по дням</CardDescription>
        </CardHeader>
        <CardContent>
          <Trend points={trend.data ?? []} />
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="font-display tracking-tight">Тепловая карта</CardTitle>
            <CardDescription>За 30 дней, день недели × час · {tz}</CardDescription>
          </CardHeader>
          <CardContent>
            <Heatmap cells={heatmap.data ?? []} />
          </CardContent>
        </Card>
        <Card className="card-hover">
          <CardHeader>
            <CardTitle className="font-display tracking-tight">По языкам</CardTitle>
            <CardDescription>Доля AI и ручного кода по языкам</CardDescription>
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
              AI-отчёт
            </CardTitle>
            <CardDescription>Сгенерирован через Gemini (или mock-режим, если ключ не задан).</CardDescription>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={() => genReport.mutate("weekly")} disabled={genReport.isPending}>
              {genReport.isPending ? "..." : "Создать недельный"}
            </Button>
            <Button size="sm" variant="outline" onClick={() => genReport.mutate("monthly")} disabled={genReport.isPending}>
              Месячный
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
            <p className="text-sm text-muted-foreground">Нажми «Создать недельный», чтобы получить первый отчёт.</p>
          )}
        </CardContent>
      </Card>

      <Card className="card-hover">
        <CardHeader>
          <CardTitle className="font-display tracking-tight">Последние события</CardTitle>
          <CardDescription>Из event store. Время в часовом поясе из настроек.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 mb-4">
            <Button onClick={() => sendDemo.mutate()} disabled={sendDemo.isPending} size="sm">
              Отправить демо-события
            </Button>
            <Button onClick={() => events.refetch()} disabled={events.isFetching} size="sm" variant="outline">
              Обновить
            </Button>
          </div>
          {eventsList.length === 0 ? (
            <EmptyState
              eyebrow="No events yet"
              title="Установи агент или отправь демо-события"
              description="Дашборд начнёт показывать данные, как только агент или браузер-расширение пришлёт первое событие."
            />
          ) : (
            <div className="overflow-x-auto rounded-md border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="py-2.5 px-3">Время</th>
                    <th className="py-2.5 px-3">Приложение</th>
                    <th className="py-2.5 px-3">Категория</th>
                    <th className="py-2.5 px-3">Источник</th>
                    <th className="py-2.5 px-3">Провайдер</th>
                    <th className="py-2.5 px-3 text-right">Длительность</th>
                  </tr>
                </thead>
                <tbody>
                  {eventsList.map((e, i) => (
                    <tr key={i} className="border-t hover:bg-muted/30 transition-colors">
                      <td className="py-2 px-3 font-mono text-xs">{formatTime(e.ts, tz)}</td>
                      <td className="py-2 px-3">{e.app_bundle}</td>
                      <td className="py-2 px-3"><CategoryBadge cat={e.category} /></td>
                      <td className="py-2 px-3 text-muted-foreground">{SOURCE_LABELS[e.source] ?? e.source}</td>
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
  return (
    <Badge tone={CATEGORY_TONES[cat] ?? "neutral"}>
      {CATEGORY_LABELS[cat] ?? cat}
    </Badge>
  );
}
