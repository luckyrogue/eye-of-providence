import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useHeatmap } from "@/entities/event";
import { getTz } from "@/shared/lib/tz";
import { reshapeHeatmap } from "../lib/reshape-heatmap";

export function HeatmapCard() {
  const { t } = useTranslation("app");
  const tz = useMemo(() => getTz(), []);
  const heat = useHeatmap(7, tz);
  const cells = useMemo(() => reshapeHeatmap(heat.data), [heat.data]);
  const hasData = cells.some((v) => v > 0);

  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-8">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.heatmap_title")}</div>
          <div className="card-sub">{t("dashboard.heatmap_sub")}</div>
        </div>
      </div>
      <div className="eop-heatmap">
        <div className="heatmap-y">
          {(
            t("dashboard.weekdays_short", {
              returnObjects: true,
              defaultValue: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
            }) as string[]
          ).map((d) => (
            <span key={d}>{d}</span>
          ))}
        </div>
        <div className="heatmap-grid">
          {cells.map((v, i) => (
            <span
              key={i}
              style={{
                background:
                  v < 0.02 ? "rgba(255,255,255,0.03)" : `hsl(13 100% 55% / ${0.12 + v * 0.7})`,
                boxShadow: v > 0.75 ? "0 0 6px hsl(13 100% 55% / 0.45)" : "none",
              }}
              title={`${Math.round(v * 60)}min`}
            />
          ))}
        </div>
      </div>
      <div className="heatmap-x">
        {["00", "04", "08", "12", "16", "20", "24"].map((h) => (
          <span key={h}>{h}</span>
        ))}
      </div>
      {!hasData && !heat.isPending && (
        <div className="text-[12px] text-muted-foreground mt-2 font-mono">
          {t("dashboard.no_data_yet")}
        </div>
      )}
    </div>
  );
}
