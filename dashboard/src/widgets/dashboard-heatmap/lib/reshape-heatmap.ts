import type { HeatmapCell } from "@/entities/event";

export function reshapeHeatmap(cells: HeatmapCell[] | undefined): number[] {
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
