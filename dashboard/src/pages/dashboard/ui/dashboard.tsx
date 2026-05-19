import { useTranslation } from "react-i18next";
import { DashboardExportEventsButton } from "./dashboard-export-events-button";
import { DashboardGenerateReportButton } from "./dashboard-generate-report-button";
import { InstallPWA } from "./install-pwa-card";
import { i18nToBcp47 } from "@/shared/lib/i18n-bcp47";
import { AttributionLog } from "@/widgets/dashboard-attribution";
import { GaugeCard } from "@/widgets/dashboard-gauge";
import { HeatmapCard } from "@/widgets/dashboard-heatmap";
import { KpiGrid } from "@/widgets/dashboard-kpi";
import { LanguageBars } from "@/widgets/dashboard-languages";
import { ProvenanceDonut } from "@/widgets/dashboard-provenance";
import { RecapCard } from "@/widgets/dashboard-recap";
import { TrendChart } from "@/widgets/dashboard-trend";

export function DashboardRoute() {
  const { t, i18n } = useTranslation("app");
  const today = new Date();
  const localeTag = i18nToBcp47(i18n.language);

  return (
    <>
      <InstallPWA />
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
          <DashboardExportEventsButton />
          <DashboardGenerateReportButton />
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
