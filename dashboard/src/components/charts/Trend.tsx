import type { TrendPoint } from "../../api/events";

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
    return <p className="text-sm text-muted-foreground">Пока нет данных за этот период.</p>;
  }

  const W = 600;
  const H = 180;
  const pad = { l: 36, r: 12, t: 8, b: 26 };
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

  // Area path для лёгкой заливки под линией manual
  const areaPath = (key: "manual" | "ai") => {
    const top = sorted
      .map(([, v], i) => {
        const x = pad.l + i * stepX;
        const y = pad.t + innerH - (v[key] / max) * innerH;
        return `${i === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");
    const lastX = pad.l + (sorted.length - 1) * stepX;
    return `${top} L ${lastX} ${pad.t + innerH} L ${pad.l} ${pad.t + innerH} Z`;
  };

  return (
    <div className="space-y-2">
      <svg viewBox={`0 0 ${W} ${H}`} className="w-full h-44">
        <defs>
          <linearGradient id="manualGrad" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="hsl(220 80% 55%)" stopOpacity="0.3" />
            <stop offset="100%" stopColor="hsl(220 80% 55%)" stopOpacity="0" />
          </linearGradient>
          <linearGradient id="aiGrad" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="hsl(280 70% 60%)" stopOpacity="0.3" />
            <stop offset="100%" stopColor="hsl(280 70% 60%)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {[0, 0.5, 1].map((t) => {
          const y = pad.t + innerH * (1 - t);
          return (
            <g key={t}>
              <line x1={pad.l} x2={W - pad.r} y1={y} y2={y} stroke="currentColor" opacity={0.08} />
              <text x={4} y={y + 3} fontSize={9} fill="currentColor" opacity={0.5}>
                {Math.round((max * t) / 60000)} мин
              </text>
            </g>
          );
        })}
        <path d={areaPath("manual")} fill="url(#manualGrad)" />
        <path d={areaPath("ai")} fill="url(#aiGrad)" />
        <path d={linePath("manual")} stroke="hsl(220 80% 55%)" strokeWidth={2} fill="none" />
        <path d={linePath("ai")} stroke="hsl(280 70% 60%)" strokeWidth={2} fill="none" />
        {sorted.map(([date], i) => {
          if (i % Math.ceil(sorted.length / 6) !== 0) return null;
          const x = pad.l + i * stepX;
          return (
            <text key={date} x={x} y={H - 8} fontSize={9} fill="currentColor" opacity={0.5} textAnchor="middle">
              {date.slice(5)}
            </text>
          );
        })}
      </svg>
      <div className="flex gap-3 text-xs text-muted-foreground">
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-3 rounded-sm" style={{ background: "hsl(220 80% 55%)" }} /> вручную
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-3 rounded-sm" style={{ background: "hsl(280 70% 60%)" }} /> AI
        </span>
      </div>
    </div>
  );
}
