// Минимальный markdown→html без лишних зависимостей.
// Покрывает то, что генерирует Gemini: # h1, ## h2, - bullets, **bold**, _italic_, `code`, абзацы.

import { useMemo } from "react";

function escape(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function inline(s: string): string {
  return escape(s)
    .replace(/`([^`]+)`/g, '<code class="rounded bg-secondary px-1 py-0.5 text-xs">$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>")
    .replace(/(^|\s)_([^_]+)_(?=\s|$)/g, "$1<em>$2</em>");
}

function render(md: string): string {
  const lines = md.split("\n");
  const out: string[] = [];
  let inList = false;
  let para: string[] = [];

  const flushPara = () => {
    if (para.length) {
      out.push(`<p class="mb-2 leading-relaxed">${inline(para.join(" "))}</p>`);
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
      out.push(`<h2 class="mt-4 mb-2 text-lg font-semibold">${inline(line.slice(3))}</h2>`);
    } else if (line.startsWith("# ")) {
      flushPara();
      closeList();
      out.push(`<h1 class="mt-2 mb-3 text-2xl font-bold">${inline(line.slice(2))}</h1>`);
    } else if (line.startsWith("- ")) {
      flushPara();
      if (!inList) {
        out.push('<ul class="mb-3 list-disc pl-5 space-y-1">');
        inList = true;
      }
      out.push(`<li>${inline(line.slice(2))}</li>`);
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
  return <div className="prose-sm" dangerouslySetInnerHTML={{ __html: html }} />;
}
