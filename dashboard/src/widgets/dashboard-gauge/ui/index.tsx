import { useTranslation } from "react-i18next";
import { useSummary } from "@/entities/event";

export function GaugeCard() {
  const { t } = useTranslation("app");
  const summary = useSummary(30);
  const ms = summary.data ?? {};
  const aiMs = Object.entries(ms)
    .filter(([k]) => k === "ai" || k.startsWith("ai_"))
    .reduce((acc, [, v]) => acc + v, 0);
  const total = aiMs + (ms["manual"] ?? ms["typed"] ?? 0);
  const value = total > 0 ? Math.round((aiMs / total) * 100) : 0;

  const r = 60;
  const cx = 80;
  const cy = 80;
  const angle = (value / 100) * 180;
  const rad = (deg: number) => ((deg - 180) * Math.PI) / 180;
  const x2 = cx + r * Math.cos(rad(angle));
  const y2 = cy + r * Math.sin(rad(angle));
  const largeArc = angle > 180 ? 1 : 0;
  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-4">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.gauge_title")}</div>
          <div className="card-sub">{t("dashboard.gauge_sub")}</div>
        </div>
      </div>
      <div className="gauge">
        <svg viewBox="0 0 160 90">
          <path
            d={`M ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx + r} ${cy}`}
            fill="none"
            stroke="rgba(255,255,255,0.05)"
            strokeWidth="10"
          />
          <path
            d={`M ${cx - r} ${cy} A ${r} ${r} 0 ${largeArc} 1 ${x2} ${y2}`}
            fill="none"
            stroke="hsl(var(--accent))"
            strokeWidth="10"
            strokeLinecap="round"
            style={{ filter: "drop-shadow(0 0 6px hsl(var(--eop-accent-glow)))" }}
          />
          <circle
            cx={x2}
            cy={y2}
            r="6"
            fill="hsl(var(--accent))"
            stroke="hsl(232 16% 10%)"
            strokeWidth="2"
          />
        </svg>
        <div className="gauge-label">
          <div className="gauge-value">
            {value}
            <span style={{ fontSize: 16, color: "hsl(var(--muted-foreground))", marginLeft: 2 }}>
              /100
            </span>
          </div>
          <div className="gauge-unit">{t("dashboard.gauge_unit")}</div>
        </div>
      </div>
      <div
        className="text-[12px] mt-3"
        style={{ color: "hsl(var(--muted-foreground))", lineHeight: 1.55 }}
      >
        {t("dashboard.gauge_lead")}
      </div>
    </div>
  );
}
