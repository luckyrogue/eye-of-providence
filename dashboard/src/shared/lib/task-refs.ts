const CU_RE = /\bCU-[a-z0-9]+\b/gi;
const HASH_RE = /(?:^|\s)#([a-z0-9]{6,12})\b/gi;
export type TaskRef = {
  id: string;
  match: string;
  url: string;
};
export function extractTaskRefs(message: string): TaskRef[] {
  const seen = new Set<string>();
  const out: TaskRef[] = [];
  let m: RegExpExecArray | null;
  while ((m = CU_RE.exec(message)) !== null) {
    const id = m[0].slice(3);
    if (seen.has(id)) continue;
    seen.add(id);
    out.push({ id, match: m[0], url: clickupURL(id) });
  }
  while ((m = HASH_RE.exec(message)) !== null) {
    const id = m[1];
    if (seen.has(id)) continue;
    seen.add(id);
    out.push({ id, match: "#" + id, url: clickupURL(id) });
  }
  return out;
}
export function clickupURL(taskID: string): string {
  return `https://app.clickup.com/t/${encodeURIComponent(taskID)}`;
}
export type Segment = {
  kind: "text" | "ref";
  value: string;
  ref?: TaskRef;
};
export function renderMessageWithLinks(message: string): Segment[] {
  const refs = extractTaskRefs(message);
  if (refs.length === 0) return [{ kind: "text", value: message }];
  const segments: Segment[] = [];
  let cursor = 0;
  const positions: {
    start: number;
    ref: TaskRef;
  }[] = [];
  for (const ref of refs) {
    let from = 0;
    while (from < message.length) {
      const idx = message.indexOf(ref.match, from);
      if (idx < 0) break;
      positions.push({ start: idx, ref });
      from = idx + ref.match.length;
    }
  }
  positions.sort((a, b) => a.start - b.start);
  for (const p of positions) {
    if (p.start < cursor) continue;
    if (p.start > cursor) {
      segments.push({ kind: "text", value: message.slice(cursor, p.start) });
    }
    segments.push({ kind: "ref", value: p.ref.match, ref: p.ref });
    cursor = p.start + p.ref.match.length;
  }
  if (cursor < message.length) {
    segments.push({ kind: "text", value: message.slice(cursor) });
  }
  return segments;
}
