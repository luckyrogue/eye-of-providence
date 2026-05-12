import { useTranslation } from "react-i18next";
import { Sparkles, Upload, Loader2 } from "lucide-react";
import { toast } from "@eop/ui";
import { useGenerateReport } from "../../entities/report";
import {
  KpiGrid,
  RecapCard,
  HeatmapCard,
  GaugeCard,
  ProvenanceDonut,
  TrendChart,
  LanguageBars,
  AttributionLog,
} from "../../widgets/dashboard-v2/ui/widgets";

export function DashboardRoute() {
  const { t, i18n } = useTranslation("app");
  const today = new Date();
  const generate = useGenerateReport();
  const localeTag = i18nToBcp47(i18n.language);

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
          <h1>{t("dashboard.overview")}</h1>
          <div className="sub font-mono" style={{ color: "hsl(var(--muted-foreground))" }}>
            {today.toLocaleDateString(localeTag, {
              month: "short",
              day: "numeric",
              weekday: "short",
            })}{" "}
            · {today.toLocaleTimeString(localeTag, { hour: "2-digit", minute: "2-digit" })}
          </div>
        </div>
        <div className="page-head-actions flex gap-2">
          {/* eslint-disable-next-line no-restricted-syntax -- custom-styled CTA (artifact 1:1) */}
          <button
            type="button"
            onClick={handleExport}
            className="inline-flex h-10 items-center gap-2 rounded-lg border px-3 text-[13px] hover:bg-foreground/5"
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
            className="btn-eop-primary inline-flex h-10 items-center gap-2 rounded-lg px-3 text-[13px] font-medium disabled:opacity-60 disabled:cursor-not-allowed"
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

      <div className="eop-grid grid grid-cols-12 gap-[14px]" style={{ marginBottom: 14 }}>
        <div className="col-span-12">
          <RecapCard />
        </div>
      </div>

      {/* Layout: Heatmap 8 + Gauge 4 (top row) → Donut 5 + Languages 7
          (middle) → Trend 12 (full width) → Attribution 12 (live).
          Удалены TimelineCard/ProvidersList/ProjectsList/FocusSessions —
          нет backend endpoints, mock data вводила в заблуждение.
          Вернутся когда добавим /v1/timeline /v1/summary/providers
          /v1/summary/projects /v1/sessions/focus. */}
      <div className="eop-grid grid grid-cols-12 gap-[14px]">
        <HeatmapCard />
        <GaugeCard />
        <ProvenanceDonut />
        <LanguageBars />
        <TrendChart />
        <AttributionLog />
      </div>
    </>
  );
}

// i18nToBcp47 — наш storage хранит "ru"/"en"/"kk"/"es", но Intl нужны BCP-47
// теги ("ru-RU", "en-US", "kk-KZ", "es-ES"). Маппим в наиболее ходовой регион.
function i18nToBcp47(lng: string): string {
  const base = lng.split("-")[0];
  switch (base) {
    case "ru":
      return "ru-RU";
    case "kk":
      return "kk-KZ";
    case "es":
      return "es-ES";
    case "en":
    default:
      return "en-US";
  }
}
