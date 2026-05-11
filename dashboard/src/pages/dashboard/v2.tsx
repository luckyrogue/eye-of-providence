// Dashboard v2 — полная реплика artifact'а (Eye of Providence (1)).
// Структура: page-head + KPI grid + Recap + 12-col grid из cards.

import { useTranslation } from "react-i18next";
import { Sparkles, Upload, Loader2 } from "lucide-react";
import { toast } from "@eop/ui";
import { useGenerateReport } from "../../entities/report";
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
  const today = new Date();
  const generate = useGenerateReport();

  const handleGenerate = () => {
    generate.mutate("weekly", {
      onSuccess: () => toast.success(t("dashboard.report_generated_toast")),
      onError: () => toast.error(t("dashboard.report_failed_toast")),
    });
  };
  const handleExport = () => {
    // TODO: backend endpoint /v1/events/export — пока нет, показываем info.
    toast.info(t("dashboard.export_not_available"));
  };
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
            onClick={handleExport}
            className="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-[13px] hover:bg-foreground/5"
            style={{ borderColor: "hsl(var(--eop-line-strong))" }}
          >
            <Upload className="h-3.5 w-3.5" />
            {t("dashboard.page_head_action_export")}
          </button>
          {/* eslint-disable-next-line no-restricted-syntax -- custom-styled CTA (artifact 1:1) */}
          <button
            type="button"
            onClick={handleGenerate}
            disabled={generate.isPending}
            className="btn-eop-primary inline-flex items-center gap-2 px-3 py-2 rounded-lg text-[13px] font-medium disabled:opacity-60 disabled:cursor-not-allowed"
          >
            {generate.isPending ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Sparkles className="h-3.5 w-3.5" />
            )}
            {t("dashboard.page_head_action_report")}
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
