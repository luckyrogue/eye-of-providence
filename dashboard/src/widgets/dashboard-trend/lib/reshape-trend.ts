import type { TrendPoint } from "@/entities/event";

export function reshapeTrend(
  points: TrendPoint[] | undefined,
): { date: string; ai: number; manual: number }[] {
  if (!points || points.length === 0) return [];
  const byDate = new Map<string, { ai: number; manual: number }>();
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
