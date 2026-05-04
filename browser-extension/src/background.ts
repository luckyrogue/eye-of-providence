// MV3 service worker — focus tracker + flush queue.
//
// Tracks: какая вкладка в фокусе, AI это или нет, длительность сессии.
// Flush: каждые 30с собирает накопленные интервалы и шлёт в backend.
//
// Privacy: НЕ отправляем заголовки страниц, URL-параметры, содержимое.
// Только host (для AI mapping) и duration.

import { aiInfoForHost } from "./ai-domains";
import { ingest, type EventPayload } from "./api";

type FocusEntry = {
  host: string;
  startedAt: number;
};

let current: FocusEntry | null = null;
const buffer: EventPayload[] = [];
const FLUSH_INTERVAL_MS = 30_000;
const IDLE_THRESHOLD_S = 90;

chrome.runtime.onInstalled.addListener(() => {
  chrome.alarms.create("flush", { periodInMinutes: FLUSH_INTERVAL_MS / 60_000 });
  chrome.idle.setDetectionInterval(IDLE_THRESHOLD_S);
});

chrome.tabs.onActivated.addListener(async ({ tabId }) => {
  try {
    const tab = await chrome.tabs.get(tabId);
    handleFocus(tab.url, tab.windowId);
  } catch {
    // tab gone
  }
});

chrome.tabs.onUpdated.addListener((_tabId, info, tab) => {
  if (info.url && tab.active) handleFocus(tab.url, tab.windowId);
});

chrome.windows.onFocusChanged.addListener(async (windowId) => {
  if (windowId === chrome.windows.WINDOW_ID_NONE) {
    closeCurrent("window-blur");
    return;
  }
  try {
    const [tab] = await chrome.tabs.query({ active: true, windowId });
    handleFocus(tab?.url, windowId);
  } catch {
    // ignore
  }
});

chrome.idle.onStateChanged.addListener((state) => {
  if (state === "idle" || state === "locked") {
    closeCurrent("idle");
  }
});

chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name === "flush") {
    closeCurrent("flush-tick"); // финализируем текущую сессию
    await flush();
  }
});

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  if (msg?.type === "ai-copy") {
    // copy event с content-script: msg.host, msg.size
    const info = aiInfoForHost(msg.host);
    if (info) {
      buffer.push({
        app_bundle: msg.host,
        category: "ai",
        source: "browser",
        ai_provider: info.provider,
        ai_channel: info.channel,
        duration_ms: 0,
        chars_in: msg.size ?? 0,
      });
    }
    sendResponse({ ok: true });
    return true;
  }
  if (msg?.type === "flush-now") {
    closeCurrent("manual-flush");
    flush().then(() => sendResponse({ ok: true }));
    return true;
  }
  return false;
});

function handleFocus(url: string | undefined, _windowId: number) {
  if (!url) {
    closeCurrent("no-url");
    return;
  }
  let host: string;
  try {
    host = new URL(url).host;
  } catch {
    return;
  }

  // Не трекаем chrome:// и подобные
  if (!url.startsWith("http")) {
    closeCurrent("non-http");
    return;
  }

  if (current && current.host === host) return;
  closeCurrent("focus-change");
  current = { host, startedAt: Date.now() };
}

function closeCurrent(_reason: string) {
  if (!current) return;
  const duration = Date.now() - current.startedAt;
  if (duration < 1000) {
    current = null;
    return;
  }
  const info = aiInfoForHost(current.host);
  if (info) {
    buffer.push({
      app_bundle: current.host,
      category: "ai",
      source: "browser",
      ai_provider: info.provider,
      ai_channel: info.channel,
      duration_ms: duration,
    });
  } else {
    // не-AI браузерное время — категория other (не скрываем сам факт)
    buffer.push({
      app_bundle: current.host,
      category: "other",
      source: "browser",
      duration_ms: duration,
    });
  }
  current = null;
}

async function flush() {
  if (buffer.length === 0) return;
  const batch = buffer.splice(0, buffer.length);
  await ingest(batch);
}
