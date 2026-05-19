import { useTranslation } from "react-i18next";
import { useSummary } from "@/entities/event";
import { formatHm } from "@/shared/lib/format-hm";
import { KpiTile } from "./kpi-tile";

export function KpiGrid() {
  const { t } = useTranslation("app");
  const summary = useSummary(7);
  const summaryPrev = useSummary(14);

  const ms = summary.data ?? {};
  const aiMs = Object.entries(ms)
    .filter(([k]) => k === "ai" || k.startsWith("ai_"))
    .reduce((acc, [, v]) => acc + v, 0);
  const manualMs = ms["manual"] ?? ms["typed"] ?? 0;
  const totalMs = aiMs + manualMs;
  const aiPct = totalMs > 0 ? Math.round((aiMs / totalMs) * 100) : 0;

  const activeFmt = formatHm(totalMs);
  const manualFmt = formatHm(manualMs);

  const prevMs = summaryPrev.data ?? {};
  const prevTotal = (prevMs["ai"] ?? 0) + (prevMs["manual"] ?? prevMs["typed"] ?? 0) - totalMs;
  const totalDelta = totalMs - prevTotal;
  const totalDeltaFmt = formatHm(Math.abs(totalDelta));

  const aiPctPrev =
    prevTotal > 0 ? Math.round((((prevMs["ai"] ?? 0) - aiMs) / prevTotal) * 100) : 0;
  const aiPctDelta = aiPct - aiPctPrev;
  const hasData = totalMs > 0;

  return (
    <div className="kpi-grid">
      <KpiTile
        label={t("dashboard.kpi_active")}
        value={activeFmt.value}
        unit={activeFmt.unit}
        delta={
          hasData && prevTotal > 0
            ? {
                kind: totalDelta >= 0 ? "up" : "down",
                text: `${totalDelta >= 0 ? "+" : "−"}${totalDeltaFmt.value}${totalDeltaFmt.unit} ${t("dashboard.kpi_vs_prior")}`,
              }
            : undefined
        }
      />
      <KpiTile
        label={t("dashboard.kpi_ai_ratio")}
        value={hasData ? String(aiPct) : "—"}
        unit={hasData ? "%" : ""}
        delta={
          hasData && prevTotal > 0
            ? {
                kind: aiPctDelta >= 0 ? "up" : "down",
                text: `${aiPctDelta >= 0 ? "+" : "−"}${Math.abs(aiPctDelta)}pp ${t("dashboard.kpi_vs_prior")}`,
              }
            : undefined
        }
      />
      <KpiTile
        label={t("dashboard.kpi_manual")}
        value={hasData ? manualFmt.value : "—"}
        unit={hasData ? manualFmt.unit : ""}
      />
    </div>
  );
}
