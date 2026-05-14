import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { useRecent, useSummary, useLanguages, useHeatmap, useTrend } from "../../../entities/event";
import { getTz } from "../../../shared/lib/tz";
import type { EventRow, HeatmapCell, LangCell, TrendPoint } from "../../../entities/event";
function formatHm(ms: number): {
  value: string;
  unit: string;
} {
  const totalMin = Math.round(ms / 60000);
  const h = Math.floor(totalMin / 60);
  const m = totalMin % 60;
  if (h > 0) return { value: String(h), unit: `h ${String(m).padStart(2, "0")}m` };
  return { value: String(m), unit: "min" };
}
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
        delta={undefined}
      />
    </div>
  );
}
function KpiTile({
  label,
  value,
  unit,
  delta,
}: {
  label: string;
  value: string;
  unit: string;
  delta?: {
    kind: "up" | "down" | "flat";
    text: string;
  };
}) {
  const Icon = delta?.kind === "up" ? TrendingUp : delta?.kind === "down" ? TrendingDown : Minus;
  return (
    <div className="kpi">
      <span className="kpi-label">{label}</span>
      <span className="kpi-value">
        {value}
        {unit && <span className="unit">{unit}</span>}
      </span>
      {delta && (
        <span className={`kpi-delta ${delta.kind}`}>
          <Icon className="h-3 w-3" />
          {delta.text}
        </span>
      )}
    </div>
  );
}
export function RecapCard() {
  const { t } = useTranslation("app");
  return (
    <div className="eop-recap">
      <div className="recap-tag">{t("dashboard.recap_tag")}</div>
      <div className="recap-text">{t("dashboard.recap_placeholder")}</div>
    </div>
  );
}
function reshapeHeatmap(cells: HeatmapCell[] | undefined): number[] {
  const out = Array.from({ length: 168 }, () => 0);
  if (!cells) return out;
  for (const c of cells) {
    if (c.dow < 0 || c.dow >= 7 || c.hour < 0 || c.hour >= 24) continue;
    const i = c.dow * 24 + c.hour;
    out[i] += c.ms;
  }
  const max = Math.max(1, ...out);
  return out.map((v) => v / max);
}
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
const PROVENANCE_BUCKETS = [
  { key: "ai_inline", labelKey: "dashboard.provenance_ai_inline", color: "hsl(var(--accent))" },
  { key: "paste_ai", labelKey: "dashboard.provenance_paste_ai", color: "#60a5fa" },
  { key: "manual", labelKey: "dashboard.provenance_manual", color: "#4ade80" },
  { key: "ai_agent", labelKey: "dashboard.provenance_ai_agent", color: "#c084fc" },
  { key: "unknown", labelKey: "dashboard.provenance_unknown", color: "rgba(255,255,255,0.18)" },
];
export function ProvenanceDonut() {
  const { t } = useTranslation("app");
  const summary = useSummary(7);
  const ms = summary.data ?? {};
  const data = PROVENANCE_BUCKETS.map((b) => ({
    ...b,
    label: t(b.labelKey),
    ms: ms[b.key] ?? (b.key === "manual" ? (ms["typed"] ?? 0) : 0),
  }));
  if (data[0].ms === 0 && ms["ai"]) {
    data[0].ms = ms["ai"];
  }
  const total = data.reduce((acc, d) => acc + d.ms, 0);
  const pcts = data.map((d) => ({
    ...d,
    pct: total > 0 ? Math.round((d.ms / total) * 100) : 0,
  }));
  const aiTotal = pcts.filter((d) => d.key.startsWith("ai_")).reduce((acc, d) => acc + d.pct, 0);
  const r = 60;
  const c = 2 * Math.PI * r;
  let offset = 0;
  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-5">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.donut_title")}</div>
          <div className="card-sub">{t("dashboard.donut_sub", { days: 7 })}</div>
        </div>
      </div>
      <div className="donut-row">
        <div className="donut">
          <svg viewBox="0 0 160 160" style={{ transform: "rotate(-90deg)" }}>
            <circle
              cx="80"
              cy="80"
              r={r}
              fill="none"
              stroke="rgba(255,255,255,0.04)"
              strokeWidth="14"
            />
            {pcts.map((s, i) => {
              const dash = (s.pct / 100) * c;
              const seg = (
                <circle
                  key={i}
                  cx="80"
                  cy="80"
                  r={r}
                  fill="none"
                  stroke={s.color}
                  strokeWidth="14"
                  strokeDasharray={`${dash} ${c - dash}`}
                  strokeDashoffset={-offset}
                  style={{ transition: "stroke-dasharray 1s ease" }}
                />
              );
              offset += dash;
              return seg;
            })}
          </svg>
          <div className="donut-center">
            <div>
              <div className="big">{aiTotal}%</div>
              <div className="lil">{t("dashboard.donut_center")}</div>
            </div>
          </div>
        </div>
        <div className="legend-row">
          {pcts.map((p) => (
            <div key={p.label} className="item">
              <span className="sw" style={{ background: p.color }} />
              <span>{p.label}</span>
              <span className="pct">{p.pct}%</span>
              <span className="lines">{Math.round(p.ms / 1000)}s</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
function reshapeTrend(points: TrendPoint[] | undefined): {
  date: string;
  ai: number;
  manual: number;
}[] {
  if (!points || points.length === 0) return [];
  const byDate = new Map<
    string,
    {
      ai: number;
      manual: number;
    }
  >();
  for (const p of points) {
    const entry = byDate.get(p.date) ?? { ai: 0, manual: 0 };
    if (p.category === "ai" || p.category.startsWith("ai_")) entry.ai += p.ms;
    else if (p.category === "manual" || p.category === "typed") entry.manual += p.ms;
    byDate.set(p.date, entry);
  }
  return Array.from(byDate.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, e]) => {
      const total = e.ai + e.manual;
      return {
        date,
        ai: total > 0 ? (e.ai / total) * 100 : 0,
        manual: total > 0 ? (e.manual / total) * 100 : 0,
      };
    });
}
export function TrendChart() {
  const { t } = useTranslation("app");
  const tz = useMemo(() => getTz(), []);
  const trend = useTrend(30, tz);
  const series = useMemo(() => reshapeTrend(trend.data), [trend.data]);
  const hasData = series.length >= 2;
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
function reshapeLanguages(cells: LangCell[] | undefined): {
  name: string;
  time: string;
  ai: number;
  manual: number;
}[] {
  if (!cells || cells.length === 0) return [];
  const byLang = new Map<
    string,
    {
      ai: number;
      manual: number;
    }
  >();
  for (const c of cells) {
    const entry = byLang.get(c.lang) ?? { ai: 0, manual: 0 };
    if (c.category === "ai" || c.category.startsWith("ai_")) entry.ai += c.ms;
    else entry.manual += c.ms;
    byLang.set(c.lang, entry);
  }
  return Array.from(byLang.entries())
    .map(([lang, e]) => {
      const total = e.ai + e.manual;
      const totalMin = Math.round(total / 60000);
      const h = Math.floor(totalMin / 60);
      const m = totalMin % 60;
      return {
        name: lang,
        time: h > 0 ? `${h}h ${String(m).padStart(2, "0")}m` : `${m}m`,
        ai: total > 0 ? Math.round((e.ai / total) * 100) : 0,
        manual: total > 0 ? Math.round((e.manual / total) * 100) : 0,
        _total: total,
      };
    })
    .sort((a, b) => b._total - a._total)
    .slice(0, 8);
}
export function LanguageBars() {
  const { t } = useTranslation("app");
  const lang = useLanguages(7);
  const langs = useMemo(() => reshapeLanguages(lang.data), [lang.data]);
  return (
    <div className="eop-card col-span-12 min-[1181px]:col-span-7">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.langs_title")}</div>
          <div className="card-sub">{t("dashboard.langs_sub")}</div>
        </div>
      </div>
      {langs.length === 0 ? (
        <div className="text-[13px] text-muted-foreground py-4">{t("dashboard.no_data_yet")}</div>
      ) : (
        <div className="langs">
          {langs.map((l) => (
            <div key={l.name} className="lang-row">
              <span className="lang-name">{l.name}</span>
              <span className="lang-time">{l.time}</span>
              <div className="lang-stack">
                <i style={{ width: `${l.ai}%`, background: "hsl(var(--accent))" }} />
                <i style={{ width: `${l.manual}%`, background: "#4ade80", opacity: 0.65 }} />
              </div>
              <span className="lang-ratio">
                {l.ai}% {t("dashboard.lang_ratio_ai")}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
function categoryToTag(c: string): string {
  if (c === "manual" || c === "typed") return "typed";
  if (c === "ai_inline" || c === "ai_assist") return "ai-inline";
  if (c === "ai_agent") return "ai-agent";
  if (c === "paste_ai") return "paste-ai";
  if (c.startsWith("refactor")) return "refactor";
  return "ai-inline";
}
function formatLogRow(r: EventRow): {
  ts: string;
  tag: string;
  file: string;
  lines: string;
  src: string;
} {
  const dt = new Date(r.ts);
  const ts = dt.toLocaleTimeString("en", { hour12: false });
  return {
    ts,
    tag: categoryToTag(r.category),
    file: r.app_bundle,
    lines: r.chars_in > 0 ? `+${r.chars_in}` : "—",
    src: r.ai_provider ? `${r.ai_provider} · ${r.ai_channel ?? r.source}` : r.source,
  };
}
export function AttributionLog() {
  const { t } = useTranslation("app");
  const recent = useRecent(20);
  const rows = useMemo(() => (recent.data ?? []).map(formatLogRow), [recent.data]);
  return (
    <div className="eop-card col-span-12">
      <div className="card-head">
        <div>
          <div className="card-title">{t("dashboard.log_title")}</div>
          <div className="card-sub">{t("dashboard.log_sub")}</div>
        </div>
        <span
          className="inline-flex items-center gap-1.5 font-mono text-[11px]"
          style={{ color: "hsl(var(--success))" }}
        >
          <span
            className="w-1.5 h-1.5 rounded-full"
            style={{
              background: "hsl(var(--success))",
              boxShadow: "0 0 6px hsl(var(--success))",
              animation: "eop-pulse 1.6s ease-in-out infinite",
            }}
          />
          LIVE
        </span>
      </div>
      <div className="log-table">
        <div className="log-row head">
          <span>{t("dashboard.log_col_time")}</span>
          <span>{t("dashboard.log_col_tag")}</span>
          <span>{t("dashboard.log_col_file")}</span>
          <span>Δ</span>
          <span />
        </div>
        {rows.length === 0 ? (
          <div className="text-[12px] text-muted-foreground py-3">{t("dashboard.no_data_yet")}</div>
        ) : (
          rows.map((r, i) => (
            <div key={i} className="log-row">
              <span className="log-time">{r.ts}</span>
              <span className={`tag ${r.tag}`}>{r.tag}</span>
              <span className="log-file truncate">
                {r.file} <span className="path">· {r.src}</span>
              </span>
              <span className="log-lines">{r.lines}</span>
              <span />
            </div>
          ))
        )}
      </div>
    </div>
  );
}
