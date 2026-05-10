// MV3 service worker — focus tracker + flush queue.
//
// MV3-особенности:
//   - Service worker может быть terminated/restarted в любой момент.
//   - In-memory state (`let current` / `let buffer`) ТЕРЯЕТСЯ при terminate.
//   - Решение: всё persistent state — в chrome.storage.session (живёт до закрытия браузера).
//
// Privacy: НЕ отправляем заголовки страниц, URL-параметры, содержимое.
// Только host (для AI mapping) и duration.

import { aiInfoForHost } from "./ai-domains";
import { ingest, type EventPayload, type IngestResult } from "./api";

type FocusEntry = {
  host: string;
  startedAt: number;
};

type StoredState = {
  current?: FocusEntry;
  buffer: EventPayload[];
  retryQueue: EventPayload[];
  retryAttempts: number;
};

const FLUSH_INTERVAL_MS = 30_000;
const IDLE_THRESHOLD_S = 90;
const MAX_RETRY_ATTEMPTS = 10; // ~10 attempts с exponential backoff = пара часов
const MAX_BUFFER_SIZE = 1000; // hard cap чтобы не утечь память при долгом offline

const STORAGE_KEY = "eop_state";

// --- Persistent state (chrome.storage.session) ---

async function loadState(): Promise<StoredState> {
  const data = await chrome.storage.session.get(STORAGE_KEY);
  return (data[STORAGE_KEY] as StoredState | undefined) ?? {
    buffer: [],
    retryQueue: [],
    retryAttempts: 0,
  };
}

async function saveState(s: StoredState): Promise<void> {
  await chrome.storage.session.set({ [STORAGE_KEY]: s });
}

async function mutate(fn: (s: StoredState) => void): Promise<void> {
  const s = await loadState();
  fn(s);
  // Hard cap — drop oldest если переполнили (offline дольше памяти).
  if (s.buffer.length > MAX_BUFFER_SIZE) {
    s.buffer = s.buffer.slice(-MAX_BUFFER_SIZE);
  }
  if (s.retryQueue.length > MAX_BUFFER_SIZE) {
    s.retryQueue = s.retryQueue.slice(-MAX_BUFFER_SIZE);
  }
  await saveState(s);
}

// --- Lifecycle ---

chrome.runtime.onInstalled.addListener(() => {
  chrome.alarms.create("flush", { periodInMinutes: FLUSH_INTERVAL_MS / 60_000 });
  chrome.idle.setDetectionInterval(IDLE_THRESHOLD_S);
});

// onStartup на каждый запуск браузера тоже создаёт alarm (на случай если
// onInstalled не сработал — например, ext был уже установлен но браузер только запустился).
chrome.runtime.onStartup.addListener(() => {
  chrome.alarms.create("flush", { periodInMinutes: FLUSH_INTERVAL_MS / 60_000 });
});

// --- Tab/window focus ---

chrome.tabs.onActivated.addListener(async ({ tabId }) => {
  try {
    const tab = await chrome.tabs.get(tabId);
    await handleFocus(tab.url);
  } catch {
    // tab gone
  }
});

chrome.tabs.onUpdated.addListener((_tabId, info, tab) => {
  if (info.url && tab.active) {
    void handleFocus(tab.url);
  }
});

chrome.windows.onFocusChanged.addListener(async (windowId) => {
  if (windowId === chrome.windows.WINDOW_ID_NONE) {
    await closeCurrent();
    return;
  }
  try {
    const [tab] = await chrome.tabs.query({ active: true, windowId });
    await handleFocus(tab?.url);
  } catch {
    // ignore
  }
});

chrome.idle.onStateChanged.addListener(async (state) => {
  if (state === "idle" || state === "locked") {
    await closeCurrent();
  }
});

chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name === "flush") {
    await closeCurrent();
    await flush();
  }
});

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  if (msg?.type === "ai-copy") {
    void handleAiCopy(msg).then(() => sendResponse({ ok: true }));
    return true;
  }
  if (msg?.type === "flush-now") {
    void closeCurrent().then(flush).then(() => sendResponse({ ok: true }));
    return true;
  }
  if (msg?.type === "pending-count") {
    void loadState().then((s) => sendResponse({ buffer: s.buffer.length, retry: s.retryQueue.length }));
    return true;
  }
  return false;
});

// --- Focus handlers ---

async function handleFocus(url: string | undefined) {
  if (!url || !url.startsWith("http")) {
    await closeCurrent();
    return;
  }
  let host: string;
  try {
    host = new URL(url).host;
  } catch {
    return;
  }

  const state = await loadState();
  if (state.current && state.current.host === host) return;

  // Финализируем предыдущий focus.
  await mutate((s) => {
    finalizeCurrent(s);
    s.current = { host, startedAt: Date.now() };
  });
}

async function closeCurrent(): Promise<void> {
  await mutate((s) => {
    finalizeCurrent(s);
    s.current = undefined;
  });
}

function finalizeCurrent(s: StoredState): void {
  if (!s.current) return;
  const duration = Date.now() - s.current.startedAt;
  if (duration < 1000) return;

  const info = aiInfoForHost(s.current.host);
  if (info) {
    s.buffer.push({
      app_bundle: s.current.host,
      category: "ai",
      source: "browser",
      ai_provider: info.provider,
      ai_channel: info.channel,
      duration_ms: duration,
    });
  } else {
    s.buffer.push({
      app_bundle: s.current.host,
      category: "other",
      source: "browser",
      duration_ms: duration,
    });
  }
}

async function handleAiCopy(msg: { host?: string; size?: number }): Promise<void> {
  if (!msg.host) return;
  const info = aiInfoForHost(msg.host);
  if (!info) return;
  await mutate((s) => {
    s.buffer.push({
      app_bundle: msg.host as string,
      category: "ai",
      source: "browser",
      ai_provider: info.provider,
      ai_channel: info.channel,
      duration_ms: 0,
      chars_in: msg.size ?? 0,
    });
  });
}

// --- Flush + retry ---

async function flush(): Promise<void> {
  const state = await loadState();
  if (state.buffer.length === 0 && state.retryQueue.length === 0) return;

  // Сначала пробуем retry-queue, потом текущий buffer.
  // Объединяем в один batch — backend всё равно держит батчи до 5000.
  const batch: EventPayload[] = [...state.retryQueue, ...state.buffer];

  // Очищаем local-state до отправки — если send fail'нёт, вернём всё в retry.
  await mutate((s) => {
    s.buffer = [];
    s.retryQueue = [];
  });

  const result: IngestResult = await ingest(batch);

  switch (result.kind) {
    case "ok":
      console.debug("[eop] flushed", batch.length, "events");
      await mutate((s) => {
        s.retryAttempts = 0;
      });
      return;

    case "no-token":
    case "client-error":
      // Дропаем batch — retry бесполезен.
      console.warn("[eop] dropped", batch.length, "events:", result.kind);
      return;

    case "retry-later": {
      // Возвращаем в retry-queue, увеличиваем attempts.
      const attempts = state.retryAttempts + 1;
      if (attempts > MAX_RETRY_ATTEMPTS) {
        console.warn("[eop] retry exhausted, dropping", batch.length, "events");
        await mutate((s) => {
          s.retryAttempts = 0;
        });
        return;
      }
      await mutate((s) => {
        // Pre-pend — старые события вперёд.
        s.retryQueue = [...batch, ...s.retryQueue];
        s.retryAttempts = attempts;
      });
      // Exponential backoff: schedule повторного flush через 2^attempts минут (cap на 30 min).
      const delayMin = Math.min(Math.pow(2, attempts), 30);
      chrome.alarms.create("retry-flush", { delayInMinutes: delayMin });
      console.log(`[eop] retry-later (#${attempts}), next flush в ${delayMin}m`);
    }
  }
}

// --- Retry alarm ---

chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name === "retry-flush") {
    await flush();
  }
});
