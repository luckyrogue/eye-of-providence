// Dashboard v2 — полная реплика artifact'а (Eye of Providence (1)).
// Структура: page-head + KPI grid + Recap + 12-col grid из cards.

import { useTranslation } from "react-i18next";
import { Sparkles, Upload } from "lucide-react";
import {
  KpiGrid,
  RecapCard,
  HeatmapCard,
  GaugeCard,
  TimelineCard,
  ProvenanceDonut,
  ProvidersList,
  TrendChart,
  LanguageBars,
  ProjectsList,
  FocusSessions,
  AttributionLog,
} from "../../widgets/dashboard-v2/ui/widgets";

export function DashboardRoute() {
  const { t } = useTranslation("app");
  void t;
  const today = new Date();
  return (
    <>
      <div className="page-head">
        <div>
          <h1>Overview</h1>
          <div className="sub font-mono" style={{ color: "hsl(var(--muted-foreground))" }}>
            {today.toLocaleDateString("en", { month: "short", day: "numeric", weekday: "short" })} ·{" "}
            {today.toLocaleTimeString("en", { hour: "2-digit", minute: "2-digit" })}
          </div>
        </div>
        <div className="page-head-actions flex gap-2">
          {/* eslint-disable-next-line no-restricted-syntax -- custom-styled CTA (artifact 1:1) */}
          <button
            type="button"
            className="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-[13px] hover:bg-foreground/5"
            style={{ borderColor: "hsl(var(--eop-line-strong))" }}
          >
            <Upload className="h-3.5 w-3.5" />
            Export CSV
          </button>
          {/* eslint-disable-next-line no-restricted-syntax -- custom-styled CTA (artifact 1:1) */}
          <button
            type="button"
            className="btn-eop-primary inline-flex items-center gap-2 px-3 py-2 rounded-lg text-[13px] font-medium"
          >
            <Sparkles className="h-3.5 w-3.5" />
            Generate weekly report
          </button>
        </div>
      </div>

      <KpiGrid />

      <div className="eop-grid" style={{ marginBottom: 14 }}>
        <div className="col-12">
          <RecapCard />
        </div>
      </div>

      <div className="eop-grid">
        <HeatmapCard />
        <GaugeCard />
        <TimelineCard />
        <ProvenanceDonut />
        <ProvidersList />
        <TrendChart />
        <LanguageBars />
        <ProjectsList />
        <FocusSessions />
        <AttributionLog />
      </div>
    </>
  );
}
