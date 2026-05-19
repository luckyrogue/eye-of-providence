import type { LangCell } from "@/entities/event";

export function reshapeLanguages(
  cells: LangCell[] | undefined,
): { name: string; time: string; ai: number; manual: number }[] {
  if (!cells || cells.length === 0) return [];
  const byLang = new Map<string, { ai: number; manual: number }>();
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
