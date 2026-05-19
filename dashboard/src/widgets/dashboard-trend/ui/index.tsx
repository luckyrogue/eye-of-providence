import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useTrend } from "@/entities/event";
import { getTz } from "@/shared/lib/tz";
import { reshapeTrend } from "../lib/reshape-trend";

export function TrendChart() {
  const { t } = useTranslation("app");
  const tz = useMemo(() => getTz(), []);
  const trend = useTrend(30, tz);
  const series = useMemo(() => reshapeTrend(trend.data), [trend.data]);
  const hasData = series.length >= 2;

  // Если данных нет — показываем empty state, БЕЗ синтетического sine wave.
  // Раньше fake sine вводил в заблуждение (выглядит как реальные метрики).
  if (!hasData) {
    return (
      <div className="eop-card col-span-12">
        <div className="card-head">
          <div>
            <div className="card-title">{t("dashboard.trend_title")}</div>
            <div className="card-sub">{t("dashboard.trend_sub")}</div>
          </div>
        </div>
        <div className="flex items-center justify-center py-12 text-[13px] text-muted-foreground">
          {t("dashboard.no_data_yet")}
        </div>
      </div>
    );
  }

  const W = 720,
    H = 180;
  const aiSeries = series.map((s) => s.ai);
  const manSeries = series.map((s) => s.manual);
  const days = aiSeries.length;
  const toPath = (s: number[]) =>
    s
      .map((v, i) => {
        const x = days === 1 ? W / 2 : (i / (days - 1)) * (W - 40) + 30;
        const y = H - 24 - (v / 100) * (H - 50);
        return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");
  const fillPath = `${toPath(aiSeries)} L ${W - 10} ${H - 24} L 30 ${H - 24} Z`;
  const avgAi = aiSeries.reduce((acc, v) => acc + v, 0) / aiSeries.length;
  const avgMan = 100 - avgAi;

  return (
    <div className="eop-card col-span-12">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.trend_title")}</div>
          <div className="card-sub">{t("dashboard.trend_sub")}</div>
        </div>
      </div>
      <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" className="w-full h-[180px]">
        <defs>
          <linearGradient id="trendGradAi" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="hsl(13 100% 55%)" stopOpacity="0.35" />
            <stop offset="100%" stopColor="hsl(13 100% 55%)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {[0, 25, 50, 75, 100].map((v) => {
          const y = H - 24 - (v / 100) * (H - 50);
          return (
            <g key={v}>
              <line x1="30" x2={W - 10} y1={y} y2={y} stroke="rgba(255,255,255,0.04)" />
              <text
                x="2"
                y={y + 3}
                fontSize="10"
                fontFamily="JetBrains Mono"
                fill="hsl(var(--muted-foreground))"
              >
                {v}%
              </text>
            </g>
          );
        })}
        <path d={fillPath} fill="url(#trendGradAi)" />
        <path
          d={toPath(aiSeries)}
          fill="none"
          stroke="hsl(var(--accent))"
          strokeWidth="2"
          style={{ filter: "drop-shadow(0 0 4px hsl(var(--eop-accent-glow)))" }}
        />
        <path d={toPath(manSeries)} fill="none" stroke="#4ade80" strokeWidth="2" />
      </svg>
      <div
        className="flex gap-4 mt-2.5 text-[12px]"
        style={{ color: "hsl(var(--muted-foreground))" }}
      >
        <span className="inline-flex items-center gap-1.5">
          <span className="inline-block w-3.5 h-0.5" style={{ background: "hsl(var(--accent))" }} />
          {t("dashboard.trend_ai")} · {Math.round(avgAi)}%
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="inline-block w-3.5 h-0.5" style={{ background: "#4ade80" }} />
          {t("dashboard.trend_manual")} · {Math.round(avgMan)}%
        </span>
      </div>
    </div>
  );
}
