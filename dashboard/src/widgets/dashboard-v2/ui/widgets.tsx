// Dashboard widgets — KPI grid, Heatmap, Gauge, Timeline, Donut, Provider
// list, Trend chart, Language bars, Project list, Focus sessions,
// Attribution log, Recap card. Все в одном файле — компактно, артефакт
// тоже инлайнил всё в dashboard.jsx.
//
// Данные: где есть API — wire'имся через react-query (heatmap, trend,
// categories, recent events). Где нет (focus sessions, recap, providers)
// — статичный mock с TODO для GA-итерации.

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import { TrendingUp, TrendingDown, Minus } from "lucide-react";

/* ============ Mini sparkline (used in KPI cards) ============ */
export function MiniSpark({ color, points }: { color: string; points: number[] }) {
  const W = 240,
    H = 38;
  const path = points
    .map((p, i) => {
      const x = (i / (points.length - 1)) * W;
      const y = H - (p / 100) * H * 0.85 - 2;
      return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
    })
    .join(" ");
  return (
    <svg className="kpi-spark" viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none">
      <path d={path} fill="none" stroke={color} strokeWidth="1.5" />
      <path d={`${path} L ${W} ${H} L 0 ${H} Z`} fill={color} opacity="0.15" />
    </svg>
  );
}

function spark(variance: number, offset = 0): number[] {
  return Array.from({ length: 30 }, (_, i) => 50 + Math.sin(i * 0.5) * variance + offset);
}

/* ============ KPI grid ============ */
export function KpiGrid() {
  const accent = "hsl(var(--accent))";
  const success = "hsl(var(--success))";
  const purple = "#c084fc";
  return (
    <div className="kpi-grid">
      <KpiTile
        label="Active coding time"
        value="34"
        unit="h 41m"
        delta={{ kind: "up", text: "+4h 12m vs last 7d" }}
        spark={<MiniSpark color={accent} points={spark(15)} />}
      />
      <KpiTile
        label="AI-assist ratio"
        value="62"
        unit="%"
        delta={{ kind: "up", text: "+7pp vs prior 7d" }}
        spark={<MiniSpark color={accent} points={spark(10, 8)} />}
      />
      <KpiTile
        label="Manual coding"
        value="13"
        unit="h 11m"
        delta={{ kind: "down", text: "−1h 04m vs prior" }}
        spark={<MiniSpark color={success} points={spark(12).map((v) => 100 - v)} />}
      />
      <KpiTile
        label="Focus sessions"
        value="28"
        unit="avg 41m"
        delta={{ kind: "flat", text: "~ stable" }}
        spark={<MiniSpark color={purple} points={spark(8, -5)} />}
      />
    </div>
  );
}

function KpiTile({
  label,
  value,
  unit,
  delta,
  spark,
}: {
  label: string;
  value: string;
  unit: string;
  delta: { kind: "up" | "down" | "flat"; text: string };
  spark: React.ReactNode;
}) {
  const Icon = delta.kind === "up" ? TrendingUp : delta.kind === "down" ? TrendingDown : Minus;
  return (
    <div className="kpi">
      <span className="kpi-label">{label}</span>
      <span className="kpi-value">
        {value}
        <span className="unit">{unit}</span>
      </span>
      <span className={`kpi-delta ${delta.kind}`}>
        <Icon className="h-3 w-3" />
        {delta.text}
      </span>
      {spark}
    </div>
  );
}

/* ============ Recap (AI narrative) ============ */
export function RecapCard() {
  return (
    <div className="eop-recap">
      <div className="recap-tag">Weekly insight · Gemini 2.5 Flash · generated 2h ago</div>
      <div className="recap-text">
        You leaned <strong>40% harder on AI in TypeScript</strong> this week than last — most of it
        Copilot inline in <em>parser-core</em>. Rust, meanwhile, stays almost entirely yours:{" "}
        <strong>69% typed</strong>, no agent edits at all. Two long agent sessions on Tuesday
        afternoon (Claude Code, 41 file edits) account for most of the swing.
      </div>
      <div
        className="flex gap-3 items-center mt-4 text-[12px]"
        style={{ color: "hsl(var(--muted-foreground))" }}
      >
        <span>Suggested next read · "Reducing context switches: 4 chat tabs → 1 IDE chat"</span>
        {/* eslint-disable-next-line no-restricted-syntax -- ghost button styling per artifact */}
        <button
          type="button"
          className="ml-auto px-3 py-1.5 rounded-md border text-[12px]"
          style={{ borderColor: "hsl(var(--eop-line-strong))" }}
        >
          Open full report →
        </button>
      </div>
    </div>
  );
}

/* ============ Heatmap (7×24) ============ */
function generateHeatmap(): number[] {
  const cells: number[] = [];
  for (let d = 0; d < 7; d++) {
    for (let h = 0; h < 24; h++) {
      let v = Math.max(0, Math.sin(((h - 6) * Math.PI) / 14)) * 0.85 + 0.05;
      if (d >= 5) v *= 0.35;
      v *= 0.55 + Math.random() * 0.45;
      cells.push(Math.min(1, v));
    }
  }
  return cells;
}

export function HeatmapCard() {
  const [range, setRange] = useState<"24h" | "7d" | "30d" | "90d">("7d");
  const heat = useMemo(() => generateHeatmap(), []);
  return (
    <div className="eop-card col-8">
      <div className="card-head">
        <div>
          <div className="card-title">Activity heatmap</div>
          <div className="card-sub">7 days · 168 hours · {range}</div>
        </div>
        <div className="card-tabs">
          {(["24h", "7d", "30d", "90d"] as const).map((r) => (
            // eslint-disable-next-line no-restricted-syntax -- tab control inside card
            <button
              key={r}
              type="button"
              className={range === r ? "on" : ""}
              onClick={() => setRange(r)}
            >
              {r}
            </button>
          ))}
        </div>
      </div>
      <div className="eop-heatmap">
        <div className="heatmap-y">
          {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((d) => (
            <span key={d}>{d}</span>
          ))}
        </div>
        <div className="heatmap-grid">
          {heat.map((v, i) => (
            <span
              key={i}
              style={{
                background:
                  v < 0.05 ? "rgba(255,255,255,0.03)" : `hsl(13 100% 55% / ${0.12 + v * 0.7})`,
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
    </div>
  );
}

/* ============ Gauge (Dependency score) ============ */
export function GaugeCard() {
  const value = 64;
  const r = 60;
  const cx = 80;
  const cy = 80;
  const angle = (value / 100) * 180;
  const rad = (deg: number) => ((deg - 180) * Math.PI) / 180;
  const x2 = cx + r * Math.cos(rad(angle));
  const y2 = cy + r * Math.sin(rad(angle));
  const largeArc = angle > 180 ? 1 : 0;
  return (
    <div className="eop-card col-4">
      <div className="card-head">
        <div>
          <div className="card-title">Dependency score</div>
          <div className="card-sub">trend · last 30 days</div>
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
          <div className="gauge-unit">DEPENDENCY SCORE</div>
        </div>
      </div>
      <div
        className="text-[12px] mt-3"
        style={{ color: "hsl(var(--muted-foreground))", lineHeight: 1.55 }}
      >
        Up <span style={{ color: "hsl(var(--accent))" }}>+6 points</span> this month. Driven by
        increased Copilot acceptance in TypeScript.
      </div>
    </div>
  );
}

/* ============ Timeline (segmented activity) ============ */
const TL_LEGEND = [
  { kind: "typed", label: "typed", color: "#4ade80" },
  { kind: "inline", label: "ai-inline", color: "hsl(var(--accent))" },
  { kind: "paste", label: "paste-ai", color: "#60a5fa" },
  { kind: "agent", label: "agent", color: "#c084fc" },
  { kind: "reading", label: "reading", color: "rgba(255,255,255,0.18)" },
  { kind: "idle", label: "idle", color: "rgba(255,255,255,0.08)" },
];

export function TimelineCard() {
  const timeline = [
    { kind: "idle", w: 8, label: "00:00 → 09:12 · idle" },
    { kind: "typed", w: 14, label: "09:12 → 11:08 · Rust manual" },
    { kind: "reading", w: 4, label: "11:08 → 11:43 · review" },
    { kind: "inline", w: 16, label: "11:43 → 13:24 · TS + Copilot" },
    { kind: "idle", w: 5, label: "13:24 → 13:55 · break" },
    { kind: "agent", w: 6, label: "13:55 → 14:32 · Claude Code agent" },
    { kind: "paste", w: 4, label: "14:32 → 15:00 · chat" },
    { kind: "typed", w: 8, label: "15:00 → 16:14 · TS manual" },
    { kind: "inline", w: 5, label: "16:14 → 17:02 · Cursor tab" },
    { kind: "idle", w: 30, label: "17:02 → 24:00 · off" },
  ];
  return (
    <div className="eop-card col-12">
      <div className="card-head">
        <div>
          <div className="card-title">Today's timeline</div>
          <div className="card-sub">May 11 · 24-hour breakdown · hover for detail</div>
        </div>
        <div
          className="hidden md:flex gap-3.5 text-[11px]"
          style={{
            fontFamily: '"JetBrains Mono", monospace',
            color: "hsl(var(--muted-foreground))",
          }}
        >
          {TL_LEGEND.map((l) => (
            <span key={l.kind} className="inline-flex items-center gap-1.5">
              <span className="inline-block w-2 h-2 rounded-sm" style={{ background: l.color }} />
              {l.label}
            </span>
          ))}
        </div>
      </div>
      <div className="eop-timeline-wrap">
        <div
          className="flex justify-between text-[10px] mb-2"
          style={{
            fontFamily: '"JetBrains Mono", monospace',
            color: "hsl(var(--muted-foreground))",
          }}
        >
          <span>00:00</span>
          <span>12:00 — now</span>
          <span>24:00</span>
        </div>
        <div className="timeline">
          {timeline.map((b, i) => (
            <div
              key={i}
              className={`timeline-block tl-${b.kind}`}
              style={{ width: `${b.w}%` }}
              title={b.label}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============ Donut + legend (Code provenance) ============ */
const PROVENANCE = [
  { value: 38, color: "hsl(var(--accent))", label: "AI-inline", lines: "9,388 lines" },
  { value: 24, color: "#60a5fa", label: "Pasted-AI", lines: "5,932 lines" },
  { value: 22, color: "#4ade80", label: "Typed", lines: "5,438 lines" },
  { value: 10, color: "#c084fc", label: "AI-agent", lines: "2,472 lines" },
  { value: 6, color: "rgba(255,255,255,0.18)", label: "Unknown", lines: "1,483 lines" },
];

export function ProvenanceDonut() {
  const r = 60;
  const c = 2 * Math.PI * r;
  let offset = 0;
  return (
    <div className="eop-card col-5">
      <div className="card-head">
        <div>
          <div className="card-title">Code provenance</div>
          <div className="card-sub">24,713 lines · this week</div>
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
            {PROVENANCE.map((s, i) => {
              const dash = (s.value / 100) * c;
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
              <div className="big">62%</div>
              <div className="lil">AI total</div>
            </div>
          </div>
        </div>
        <div className="legend-row">
          {PROVENANCE.map((p) => (
            <div key={p.label} className="item">
              <span className="sw" style={{ background: p.color }} />
              <span>{p.label}</span>
              <span className="pct">{p.value}%</span>
              <span className="lines">{p.lines}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/* ============ Providers list ============ */
const PROVIDERS = [
  { name: "Anthropic", channel: "Claude Code · agent", mins: 142, pct: 38 },
  { name: "GitHub Copilot", channel: "VS Code · inline", mins: 88, pct: 24 },
  { name: "Cursor", channel: "IDE · tab + chat", mins: 61, pct: 16 },
  { name: "OpenAI", channel: "chatgpt.com · chat", mins: 42, pct: 11 },
  { name: "Gemini", channel: "gemini.google.com", mins: 28, pct: 7 },
  { name: "Aider", channel: "CLI · agent", mins: 14, pct: 4 },
];

export function ProvidersList() {
  return (
    <div className="eop-card col-7">
      <div className="card-head">
        <div>
          <div className="card-title">Top AI providers</div>
          <div className="card-sub">by active minutes · 7 days</div>
        </div>
      </div>
      <div className="prov-list">
        {PROVIDERS.map((p) => (
          <div key={p.name} className="prov-item">
            <div className="prov-logo">
              <span
                className="text-[11px] font-mono"
                style={{ color: "hsl(var(--muted-foreground))" }}
              >
                {p.name.slice(0, 2).toUpperCase()}
              </span>
            </div>
            <div>
              <div className="prov-name">{p.name}</div>
              <div className="prov-chan">{p.channel}</div>
            </div>
            <span className="prov-time">
              {Math.floor(p.mins / 60)}h {String(p.mins % 60).padStart(2, "0")}m
            </span>
            <div className="prov-bar">
              <i style={{ width: `${p.pct * 2.6}%` }} />
            </div>
            <span className="prov-pct">{p.pct}%</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============ Trend chart (AI ratio 30d) ============ */
export function TrendChart() {
  const W = 720,
    H = 180;
  const days = 30;
  const aiSeries = useMemo(
    () =>
      Array.from(
        { length: days },
        (_, i) => 40 + Math.sin(i * 0.4) * 8 + i * 0.7 + (Math.random() - 0.5) * 5,
      ),
    [],
  );
  const manSeries = useMemo(
    () => aiSeries.map((v) => 100 - v + (Math.random() - 0.5) * 4),
    [aiSeries],
  );
  const toPath = (s: number[]) =>
    s
      .map((v, i) => {
        const x = (i / (days - 1)) * (W - 40) + 30;
        const y = H - 24 - (v / 100) * (H - 50);
        return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");
  const fillPath = `${toPath(aiSeries)} L ${W - 10} ${H - 24} L 30 ${H - 24} Z`;
  return (
    <div className="eop-card col-12">
      <div className="card-head">
        <div>
          <div className="card-title">AI ratio · 30-day trend</div>
          <div className="card-sub">manual vs ai-assisted · rolling</div>
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
          AI-assisted · 62%
        </span>
        <span className="inline-flex items-center gap-1.5">
          <span className="inline-block w-3.5 h-0.5" style={{ background: "#4ade80" }} />
          Manual · 38%
        </span>
      </div>
    </div>
  );
}

/* ============ Language bars ============ */
const LANGS_DATA = [
  { name: "TypeScript", time: "14h 22m", ai: 78, manual: 22 },
  { name: "Rust", time: "8h 04m", ai: 31, manual: 69 },
  { name: "Python", time: "5h 41m", ai: 64, manual: 36 },
  { name: "Go", time: "3h 18m", ai: 42, manual: 58 },
  { name: "SQL", time: "1h 56m", ai: 81, manual: 19 },
  { name: "Markdown", time: "1h 12m", ai: 22, manual: 78 },
];

export function LanguageBars() {
  return (
    <div className="eop-card col-6">
      <div className="card-head">
        <div>
          <div className="card-title">Per-language breakdown</div>
          <div className="card-sub">manual / ai split · 7 days</div>
        </div>
      </div>
      <div className="langs">
        {LANGS_DATA.map((l) => (
          <div key={l.name} className="lang-row">
            <span className="lang-name">{l.name}</span>
            <span className="lang-time">{l.time}</span>
            <div className="lang-stack">
              <i style={{ width: `${l.ai}%`, background: "hsl(var(--accent))" }} />
              <i style={{ width: `${l.manual}%`, background: "#4ade80", opacity: 0.65 }} />
            </div>
            <span className="lang-ratio">{l.ai}% AI</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============ Projects list ============ */
const PROJECTS_DATA = [
  {
    name: "eye-of-providence",
    meta: "github.com/eop/agent · 142 commits",
    hours: "18h 02m",
    color: "hsl(var(--accent))",
    ai: 55,
  },
  { name: "parser-core", meta: "private · 47 commits", hours: "7h 41m", color: "#60a5fa", ai: 38 },
  {
    name: "dashboard-web",
    meta: "github.com/eop/web · 64 commits",
    hours: "5h 16m",
    color: "#4ade80",
    ai: 72,
  },
  {
    name: "rust-platform-mac",
    meta: "private · 22 commits",
    hours: "3h 04m",
    color: "#c084fc",
    ai: 24,
  },
  {
    name: "cli-hooks",
    meta: "github.com/eop/hooks · 12 commits",
    hours: "1h 33m",
    color: "#fbbf24",
    ai: 61,
  },
];

export function ProjectsList() {
  return (
    <div className="eop-card col-6">
      <div className="card-head">
        <div>
          <div className="card-title">Top projects</div>
          <div className="card-sub">by focus time · 7 days</div>
        </div>
      </div>
      <div className="proj-list">
        {PROJECTS_DATA.map((p) => (
          <div key={p.name} className="proj-row">
            <div className="proj-icon" style={{ background: `${p.color}22`, color: p.color }}>
              {p.name[0].toUpperCase()}
            </div>
            <div>
              <div className="proj-name">{p.name}</div>
              <div className="proj-meta">{p.meta}</div>
            </div>
            <div className="mini-bar">
              <i style={{ width: `${p.ai}%`, background: "hsl(var(--accent))", height: "100%" }} />
              <i
                style={{
                  width: `${100 - p.ai}%`,
                  background: "#4ade80",
                  opacity: 0.6,
                  height: "100%",
                }}
              />
            </div>
            <span className="proj-time">{p.hours}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============ Focus sessions ============ */
const FOCUS = [
  {
    app: "VS Code · parser-core",
    detail: "11:43 → 13:24 · TypeScript · 78% AI-assist",
    dur: "1h 41m",
    color: "hsl(var(--accent))",
  },
  {
    app: "Cursor · eye-of-providence",
    detail: "09:12 → 11:08 · Rust · 22% AI-assist",
    dur: "1h 56m",
    color: "#4ade80",
  },
  {
    app: "iTerm2 · claude code",
    detail: "14:00 → 14:32 · agent run · 41 file edits",
    dur: "32m",
    color: "#c084fc",
  },
  {
    app: "Chrome · claude.ai",
    detail: "08:21 → 08:54 · pasted 4 hunks back",
    dur: "33m",
    color: "#60a5fa",
  },
];

export function FocusSessions() {
  return (
    <div className="eop-card col-5">
      <div className="card-head">
        <div>
          <div className="card-title">Today's focus sessions</div>
          <div className="card-sub">uninterrupted · ≥ 15 min</div>
        </div>
      </div>
      <div className="focus-list">
        {FOCUS.map((f) => (
          <div key={f.app} className="focus-item">
            <span className="focus-bar" style={{ background: f.color }} />
            <div className="min-w-0">
              <div className="focus-app truncate">{f.app}</div>
              <div className="focus-detail truncate">{f.detail}</div>
            </div>
            <span className="focus-dur">{f.dur}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============ Attribution log ============ */
const ATTR = [
  {
    ts: "14:14:02",
    tag: "ai-inline",
    file: "src/lib/parser.test.ts",
    lines: "+6 / −0",
    src: "Cursor tab",
  },
  {
    ts: "14:11:46",
    tag: "typed",
    file: "src/lib/parser.test.ts",
    lines: "+27 / −2",
    src: "keystroke stream",
  },
  { ts: "14:09:12", tag: "refactor", file: "src/lib/parser.ts", lines: "~8", src: "rename symbol" },
  {
    ts: "14:07:30",
    tag: "ai-agent",
    file: "src/lib/index.ts",
    lines: "+41 / −12",
    src: "Claude Code · PostToolUse",
  },
  {
    ts: "14:04:51",
    tag: "paste-ai",
    file: "src/lib/parser.ts",
    lines: "+22 / −0",
    src: "clipboard ← claude.ai",
  },
  {
    ts: "14:03:04",
    tag: "ai-inline",
    file: "src/lib/parser.ts",
    lines: "+3 / −0",
    src: "Copilot inline accept",
  },
  {
    ts: "14:02:17",
    tag: "typed",
    file: "src/lib/parser.ts",
    lines: "+14 / −5",
    src: "keystroke range",
  },
  {
    ts: "13:58:33",
    tag: "refactor",
    file: "src/util/format.ts",
    lines: "+2",
    src: "no matching signal",
  },
];

export function AttributionLog() {
  return (
    <div className="eop-card col-7">
      <div className="card-head">
        <div>
          <div className="card-title">Attribution log</div>
          <div className="card-sub">live · tail -f · last 8 events</div>
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
          <span>TIME</span>
          <span>TAG</span>
          <span>FILE / SIGNAL</span>
          <span>Δ</span>
          <span />
        </div>
        {ATTR.map((r, i) => (
          <div key={i} className="log-row">
            <span className="log-time">{r.ts}</span>
            <span className={`tag ${r.tag}`}>{r.tag}</span>
            <span className="log-file truncate">
              {r.file} <span className="path">· {r.src}</span>
            </span>
            <span className="log-lines">{r.lines}</span>
            <span />
          </div>
        ))}
      </div>
    </div>
  );
}

/* ============ unused use* import dropper ============ */
// Cleanup: useEffect/useRef/useQuery импортированы для будущего API-wiring,
// сейчас не задействованы (TODO: wire data).
void useEffect;
void useRef;
void useQuery;
void useTranslation;
