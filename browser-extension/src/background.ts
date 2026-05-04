// MV3 service worker — event hub.
// Phase 2: tabs/windows focus tracking, AI domain detection, events → local agent.

const AI_DOMAINS = new Set([
  "chatgpt.com",
  "chat.openai.com",
  "claude.ai",
  "gemini.google.com",
  "copilot.microsoft.com",
  "www.perplexity.ai",
  "poe.com",
  "you.com",
  "cursor.sh",
  "www.phind.com",
  "bolt.new",
  "v0.dev",
  "lovable.dev",
]);

chrome.tabs.onActivated.addListener(async (info) => {
  const tab = await chrome.tabs.get(info.tabId);
  const host = tab.url ? new URL(tab.url).host : "";
  console.debug("[eop] tab activated", { host, isAi: AI_DOMAINS.has(host) });
  // TODO Phase 2: send focus_change event to local agent (127.0.0.1:PORT).
});

chrome.windows.onFocusChanged.addListener((windowId) => {
  console.debug("[eop] window focus", windowId);
  // TODO Phase 2: track browser-level focus loss (assume idle while not focused).
});
