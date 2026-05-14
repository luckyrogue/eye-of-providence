import { useTranslation } from "react-i18next";
import { Sparkles, Upload, Loader2 } from "lucide-react";
import { Button, toast } from "@eop/ui";
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
          <Button
            type="button"
            variant="outline"
            onClick={handleExport}
            className="inline-flex h-10 items-center gap-2 rounded-lg px-3 text-[13px]"
            style={{ borderColor: "hsl(var(--eop-line-strong))" }}
          >
            <Upload className="h-3.5 w-3.5" />
            {t("dashboard.page_head_action_export")}
          </Button>
          <Button
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
          </Button>
        </div>
      </div>

      <KpiGrid />

      <div className="eop-grid grid grid-cols-12 gap-[14px]" style={{ marginBottom: 14 }}>
        <div className="col-span-12">
          <RecapCard />
        </div>
      </div>

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
