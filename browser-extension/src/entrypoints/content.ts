import { aiInfoForHost } from "../shared/api/ai-domains";
import type { AiCopyMessage } from "../shared/api/messages";
const host = location.host;
const info = aiInfoForHost(host);
if (info) {
  document.addEventListener("copy", onCopy, true);
}
function onCopy(_event: ClipboardEvent) {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0) return;
  const range = sel.getRangeAt(0);
  if (!isInsideAssistantMessage(range.commonAncestorContainer)) return;
  const size = sel.toString().length;
  if (size === 0) return;
  const msg: AiCopyMessage = { type: "ai-copy", host, size };
  chrome.runtime.sendMessage(msg).catch(() => {});
}
function isInsideAssistantMessage(node: Node): boolean {
  let el: Element | null = node.nodeType === 1 ? (node as Element) : node.parentElement;
  while (el) {
    if (matchesAssistant(el)) return true;
    el = el.parentElement;
  }
  return false;
}
function matchesAssistant(el: Element): boolean {
  if (el.getAttribute?.("data-message-author-role") === "assistant") return true;
  const role = el.getAttribute?.("data-author-role") || el.getAttribute?.("data-role");
  if (role === "assistant") return true;
  const cls = (el.className || "").toString().toLowerCase();
  if (cls.includes("assistant") || cls.includes("agent-message") || cls.includes("ai-message"))
    return true;
  return false;
}
