// Heatmap 7 (DOW) × 24 (hour). Цвет — суммарная активность за всё категории.
// Tooltip показывает breakdown по категориям.

import { useMemo, useState } from "react";
import type { HeatmapCell } from "./api";

const DAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

export function Heatmap({ cells }: { cells: HeatmapCell[] }) {
  const [hover, setHover] = useState<{ dow: number; hour: number } | null>(null);

  const grid = useMemo(() => buildGrid(cells), [cells]);
  const max = useMemo(() => Math.max(1, ...grid.flat().map((c) => c.total)), [grid]);

  const hovered = hover ? grid[hover.dow][hover.hour] : null;

  return (
    <div className="space-y-2">
      <div className="overflow-x-auto">
        <div className="inline-block">
          <div className="grid" style={{ gridTemplateColumns: "auto repeat(24, 1.25rem)" }}>
            <div></div>
            {Array.from({ length: 24 }, (_, h) => (
              <div key={h} className="text-center text-[10px] text-muted-foreground tabular-nums">
                {h % 3 === 0 ? h : ""}
              </div>
            ))}
            {DAYS.map((day, dow) => (
              <DayRow
                key={day}
                day={day}
                row={grid[dow]}
                max={max}
                onHover={(hour) => setHover({ dow, hour })}
                onLeave={() => setHover(null)}
              />
            ))}
          </div>
        </div>
      </div>
      <div className="min-h-[2.5rem] rounded-md border bg-card p-2 text-xs">
        {hovered && hovered.total > 0 ? (
          <div className="flex items-center gap-3">
            <span className="font-mono">
              {DAYS[hover!.dow]} {String(hover!.hour).padStart(2, "0")}:00
            </span>
            <span>
              total: <strong>{Math.round(hovered.total / 60)}m</strong>
            </span>
            {hovered.byCat.map(({ cat, ms }) => (
              <span key={cat} className="text-muted-foreground">
                {cat}: {Math.round(ms / 60)}m
              </span>
            ))}
          </div>
        ) : (
          <span className="text-muted-foreground">Hover over a cell to see breakdown.</span>
        )}
      </div>
    </div>
  );
}

type GridCell = { total: number; byCat: { cat: string; ms: number }[] };

function buildGrid(cells: HeatmapCell[]): GridCell[][] {
  const grid: GridCell[][] = Array.from({ length: 7 }, () =>
    Array.from({ length: 24 }, () => ({ total: 0, byCat: [] })),
  );
  for (const c of cells) {
    if (c.dow < 0 || c.dow > 6 || c.hour < 0 || c.hour > 23) continue;
    const cell = grid[c.dow][c.hour];
    cell.total += c.ms / 1000;
    cell.byCat.push({ cat: c.category, ms: c.ms / 1000 });
  }
  for (const row of grid) for (const cell of row) cell.byCat.sort((a, b) => b.ms - a.ms);
  return grid;
}

function DayRow({
  day,
  row,
  max,
  onHover,
  onLeave,
}: {
  day: string;
  row: GridCell[];
  max: number;
  onHover: (hour: number) => void;
  onLeave: () => void;
}) {
  return (
    <>
      <div className="pr-2 text-xs text-muted-foreground tabular-nums">{day}</div>
      {row.map((cell, hour) => {
        const intensity = Math.min(1, cell.total / max);
        const bg =
          cell.total === 0
            ? "hsl(var(--secondary))"
            : `hsla(220, 80%, 50%, ${0.15 + intensity * 0.8})`;
        return (
          <div
            key={hour}
            className="h-5 w-5 rounded-sm border border-border/40"
            style={{ background: bg }}
            onMouseEnter={() => onHover(hour)}
            onMouseLeave={onLeave}
          />
        );
      })}
    </>
  );
}
