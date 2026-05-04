// AI vs manual trend per day. Простой SVG-line chart без зависимостей.

import type { TrendPoint } from "./api";

export function Trend({ points }: { points: TrendPoint[] }) {
  const byDate = new Map<string, { manual: number; ai: number }>();
  for (const p of points) {
    const v = byDate.get(p.date) ?? { manual: 0, ai: 0 };
    if (p.category === "ai") v.ai += p.ms;
    else if (p.category === "manual" || p.category === "refactor") v.manual += p.ms;
    byDate.set(p.date, v);
  }
  const sorted = Array.from(byDate.entries()).sort(([a], [b]) => (a < b ? -1 : 1));
  if (sorted.length === 0) {
    return <p className="text-sm text-muted-foreground">Нет данных за этот период.</p>;
  }

  const W = 600;
  const H = 160;
  const pad = { l: 30, r: 12, t: 8, b: 22 };
  const innerW = W - pad.l - pad.r;
  const innerH = H - pad.t - pad.b;
  const max = Math.max(1, ...sorted.flatMap(([, v]) => [v.manual, v.ai]));
  const stepX = sorted.length > 1 ? innerW / (sorted.length - 1) : innerW;

  const linePath = (key: "manual" | "ai") =>
    sorted
      .map(([, v], i) => {
        const x = pad.l + i * stepX;
        const y = pad.t + innerH - (v[key] / max) * innerH;
        return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");

  return (
    <div className="space-y-1">
      <svg viewBox={`0 0 ${W} ${H}`} className="w-full h-40">
        {/* y axis ticks */}
        {[0, 0.5, 1].map((t) => {
          const y = pad.t + innerH * (1 - t);
          return (
            <g key={t}>
              <line x1={pad.l} x2={W - pad.r} y1={y} y2={y} stroke="currentColor" opacity={0.1} />
              <text x={4} y={y + 3} fontSize={9} fill="currentColor" opacity={0.5}>
                {Math.round((max * t) / 60000)}m
              </text>
            </g>
          );
        })}
        <path d={linePath("manual")} stroke="hsl(220 80% 55%)" strokeWidth={2} fill="none" />
        <path d={linePath("ai")} stroke="hsl(280 70% 60%)" strokeWidth={2} fill="none" />
        {/* x axis dates */}
        {sorted.map(([date], i) => {
          if (i % Math.ceil(sorted.length / 6) !== 0) return null;
          const x = pad.l + i * stepX;
          return (
            <text key={date} x={x} y={H - 6} fontSize={9} fill="currentColor" opacity={0.5} textAnchor="middle">
              {date.slice(5)}
            </text>
          );
        })}
      </svg>
      <div className="flex gap-3 text-xs text-muted-foreground">
        <span className="flex items-center gap-1">
          <span className="h-2 w-3 rounded-sm" style={{ background: "hsl(220 80% 55%)" }} /> manual
        </span>
        <span className="flex items-center gap-1">
          <span className="h-2 w-3 rounded-sm" style={{ background: "hsl(280 70% 60%)" }} /> ai
        </span>
      </div>
    </div>
  );
}
