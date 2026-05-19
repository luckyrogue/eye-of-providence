import type { EventRow } from "@/entities/event";

function categoryToTag(c: string): string {
  if (c === "manual" || c === "typed") return "typed";
  if (c === "ai_inline" || c === "ai_assist") return "ai-inline";
  if (c === "ai_agent") return "ai-agent";
  if (c === "paste_ai") return "paste-ai";
  if (c.startsWith("refactor")) return "refactor";
  return "ai-inline";
}

export function formatLogRow(r: EventRow): {
  ts: string;
  tag: string;
  file: string;
  lines: string;
  src: string;
} {
  const dt = new Date(r.ts);
  const ts = dt.toLocaleTimeString("en", { hour12: false });
  return {
    ts,
    tag: categoryToTag(r.category),
    file: r.app_bundle,
    lines: r.chars_in > 0 ? `+${r.chars_in}` : "—",
    src: r.ai_provider ? `${r.ai_provider} · ${r.ai_channel ?? r.source}` : r.source,
  };
}
