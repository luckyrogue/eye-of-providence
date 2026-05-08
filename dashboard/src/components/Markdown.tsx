// Минимальный markdown→html без зависимостей.
// Покрывает то, что генерирует Gemini: # h1, ## h2, - bullets, **bold**, _italic_, `code`, абзацы.
// Плюс: подсветка чисел/процентов в bold (badge-style) и иконок в начале заголовков.

import { useMemo } from "react";

function escape(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function inline(s: string): string {
  let r = escape(s);
  // `inline code` — фон + моно
  r = r.replace(/`([^`]+)`/g, '<code class="rounded bg-purple-500/10 px-1.5 py-0.5 text-[0.85em] font-mono text-purple-700 dark:text-purple-300">$1</code>');
  // **bold** — выделяем числа/проценты яркой подсветкой, остальное — просто bold
  r = r.replace(/\*\*([^*]+)\*\*/g, (_, t: string) => {
    const isNumeric = /^-?[\d.,]+\s*(%|сек|мин|ч|секунд|минут|часов|часа|симв|событий)?$|^-?[\d.,]+%/.test(t);
    if (isNumeric) {
      return `<strong class="text-foreground bg-amber-500/10 rounded px-1 py-0.5">${t}</strong>`;
    }
    return `<strong class="font-semibold text-foreground">${t}</strong>`;
  });
  // _italic_
  r = r.replace(/(^|\s)_([^_]+)_(?=\s|$|[.,!?])/g, '$1<em class="text-muted-foreground">$2</em>');
  // ▸ → spaced bullet
  r = r.replace(/▸/g, '<span class="text-purple-500">▸</span>');
  return r;
}

function render(md: string): string {
  const lines = md.split("\n");
  const out: string[] = [];
  let inList = false;
  let para: string[] = [];

  const flushPara = () => {
    if (para.length) {
      out.push(`<p class="mb-3 leading-relaxed text-foreground/90">${inline(para.join(" "))}</p>`);
      para = [];
    }
  };
  const closeList = () => {
    if (inList) {
      out.push("</ul>");
      inList = false;
    }
  };

  for (const raw of lines) {
    const line = raw.trimEnd();
    if (!line.trim()) {
      flushPara();
      closeList();
      continue;
    }
    if (line.startsWith("## ")) {
      flushPara();
      closeList();
      const text = line.slice(3);
      out.push(`<h2 class="mt-6 mb-3 text-lg font-semibold tracking-tight flex items-center gap-2">${inline(text)}</h2>`);
    } else if (line.startsWith("# ")) {
      flushPara();
      closeList();
      const text = line.slice(2);
      out.push(`<h1 class="mt-2 mb-4 text-2xl font-bold tracking-tight">${inline(text)}</h1>`);
    } else if (/^\s*-\s+/.test(line)) {
      flushPara();
      if (!inList) {
        out.push('<ul class="mb-4 space-y-1.5 pl-1">');
        inList = true;
      }
      const item = line.replace(/^\s*-\s+/, "");
      out.push(`<li class="flex gap-2 text-foreground/90"><span class="text-purple-500/60 mt-1">•</span><span class="flex-1">${inline(item)}</span></li>`);
    } else {
      closeList();
      para.push(line);
    }
  }
  flushPara();
  closeList();
  return out.join("\n");
}

export function Markdown({ source }: { source: string }) {
  const html = useMemo(() => render(source), [source]);
  return <div className="prose-eop" dangerouslySetInnerHTML={{ __html: html }} />;
}
