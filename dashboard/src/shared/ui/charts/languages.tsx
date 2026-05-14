type LangCell = {
  lang: string;
  category: string;
  chars: number;
};
export function Languages({ cells, top = 6 }: { cells: LangCell[]; top?: number }) {
  const byLang = new Map<
    string,
    {
      manual: number;
      ai: number;
      refactor: number;
      other: number;
      total: number;
    }
  >();
  for (const c of cells) {
    const v = byLang.get(c.lang) ?? { manual: 0, ai: 0, refactor: 0, other: 0, total: 0 };
    if (c.category === "manual") v.manual += c.chars;
    else if (c.category === "ai") v.ai += c.chars;
    else if (c.category === "refactor") v.refactor += c.chars;
    else v.other += c.chars;
    v.total += c.chars;
    byLang.set(c.lang, v);
  }
  const sorted = Array.from(byLang.entries())
    .sort((a, b) => b[1].total - a[1].total)
    .slice(0, top);
  if (sorted.length === 0) {
    return <p className="text-sm text-muted-foreground">Нет данных по языкам за этот период.</p>;
  }
  const max = Math.max(1, ...sorted.map(([, v]) => v.total));
  return (
    <div className="space-y-3">
      {sorted.map(([lang, v]) => {
        const aiPct = v.total ? Math.round((v.ai / v.total) * 100) : 0;
        const manualPct = v.total ? Math.round((v.manual / v.total) * 100) : 0;
        const widthPct = (v.total / max) * 100;
        return (
          <div key={lang} className="space-y-1.5">
            <div className="flex items-baseline justify-between text-sm">
              <div className="font-medium">{lang}</div>
              <div className="text-xs text-muted-foreground tabular-nums">
                {v.total.toLocaleString("ru-RU")} симв · AI {aiPct}% · вручную {manualPct}%
              </div>
            </div>
            <div
              className="flex h-2.5 overflow-hidden rounded-full bg-secondary"
              style={{ width: `${widthPct}%` }}
            >
              {v.manual > 0 && (
                <div className="bg-blue-500" style={{ width: `${(v.manual / v.total) * 100}%` }} />
              )}
              {v.ai > 0 && (
                <div className="bg-purple-500" style={{ width: `${(v.ai / v.total) * 100}%` }} />
              )}
              {v.refactor > 0 && (
                <div
                  className="bg-amber-500"
                  style={{ width: `${(v.refactor / v.total) * 100}%` }}
                />
              )}
              {v.other > 0 && (
                <div
                  className="bg-neutral-400"
                  style={{ width: `${(v.other / v.total) * 100}%` }}
                />
              )}
            </div>
          </div>
        );
      })}
      <div className="flex flex-wrap gap-3 pt-2 text-xs text-muted-foreground">
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm bg-blue-500" /> вручную
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm bg-purple-500" /> AI
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm bg-amber-500" /> рефакторинг
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-sm bg-neutral-400" /> прочее
        </span>
      </div>
    </div>
  );
}
