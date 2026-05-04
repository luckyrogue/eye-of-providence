// Content script — слушает copy events и определяет, был ли источник
// сообщением AI-ассистента. Только если "да" — шлёт в background.
//
// Privacy: содержимое сообщений НЕ передаётся, только размер и host.

import { aiInfoForHost } from "./ai-domains";

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

  chrome.runtime.sendMessage({ type: "ai-copy", host, size }).catch(() => {
    // service worker возможно ещё не готов — это нормально
  });
}

// Эвристика: проверяем ancestor chain на маркеры AI-сообщений.
// Селекторы подобраны вручную для основных провайдеров. Phase 3 — расширить.
function isInsideAssistantMessage(node: Node): boolean {
  let el: Element | null = node.nodeType === 1 ? (node as Element) : node.parentElement;
  while (el) {
    if (matchesAssistant(el)) return true;
    el = el.parentElement;
  }
  return false;
}

function matchesAssistant(el: Element): boolean {
  // ChatGPT: data-message-author-role="assistant"
  if (el.getAttribute?.("data-message-author-role") === "assistant") return true;
  // Claude: data-testid="message" + assistant role
  const role = el.getAttribute?.("data-author-role") || el.getAttribute?.("data-role");
  if (role === "assistant") return true;
  // Generic class hint
  const cls = (el.className || "").toString().toLowerCase();
  if (cls.includes("assistant") || cls.includes("agent-message") || cls.includes("ai-message")) return true;
  return false;
}
