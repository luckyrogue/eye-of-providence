// MV3 service worker — focus tracker + flush queue.
//
// MV3-особенности:
//   - Service worker может быть terminated/restarted в любой момент.
//   - In-memory state (`let current` / `let buffer`) ТЕРЯЕТСЯ при terminate.
//   - Решение: persistent state — в chrome.storage.local (переживает
//     рестарт браузера и переустановку расширения; events не теряются
//     при offline > сессии).
//
// Privacy: НЕ отправляем заголовки страниц, URL-параметры, содержимое.
// Только host (для AI mapping) и duration.

import { aiInfoForHost } from "../shared/api/ai-domains";
import { ingest, type EventPayload, type IngestResult } from "../shared/api/backend";
import type {
  AiCopyResponse,
  ExtensionMessage,
  FlushNowResponse,
  GetStatusResponse,
  PendingCountResponse,
  SetPausedResponse,
} from "../shared/api/messages";

type FocusEntry = {
  host: string;
  startedAt: number;
};

type StoredState = {
  current?: FocusEntry;
  buffer: EventPayload[];
  retryQueue: EventPayload[];
  retryAttempts: number;
  paused: boolean;
  lastSuccessTs: number;
};

const FLUSH_INTERVAL_MS = 30_000;
const IDLE_THRESHOLD_S = 90;
const MAX_RETRY_ATTEMPTS = 10; // ~10 attempts с exponential backoff = пара часов
const MAX_BUFFER_SIZE = 1000; // hard cap чтобы не утечь память при долгом offline

const STORAGE_KEY = "eop_state";

async function loadState(): Promise<StoredState> {
  const data = await chrome.storage.local.get(STORAGE_KEY);
  const raw = data[STORAGE_KEY] as Partial<StoredState> | undefined;
  return {
    current: raw?.current,
    buffer: raw?.buffer ?? [],
    retryQueue: raw?.retryQueue ?? [],
    retryAttempts: raw?.retryAttempts ?? 0,
    paused: raw?.paused ?? false,
    lastSuccessTs: raw?.lastSuccessTs ?? 0,
  };
}

async function saveState(s: StoredState): Promise<void> {
  await chrome.storage.local.set({ [STORAGE_KEY]: s });
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

chrome.runtime.onInstalled.addListener((details) => {
  void chrome.alarms.create("flush", { periodInMinutes: FLUSH_INTERVAL_MS / 60_000 });
  chrome.idle.setDetectionInterval(IDLE_THRESHOLD_S);
  if (details.reason === "install") {
    // Открываем onboarding-страницу один раз при установке.
    void chrome.tabs.create({ url: chrome.runtime.getURL("options.html") });
  }
});

// onStartup на каждый запуск браузера тоже создаёт alarm (на случай если
// onInstalled не сработал — например, ext был уже установлен но браузер только запустился).
chrome.runtime.onStartup.addListener(() => {
  void chrome.alarms.create("flush", { periodInMinutes: FLUSH_INTERVAL_MS / 60_000 });
});

chrome.tabs.onActivated.addListener(({ tabId }) => {
  void (async () => {
    try {
      const tab = await chrome.tabs.get(tabId);
      await handleFocus(tab.url);
    } catch {
      // tab gone
    }
  })();
});

chrome.tabs.onUpdated.addListener((_tabId, info, tab) => {
  if (info.url && tab.active) {
    void handleFocus(tab.url);
  }
});

chrome.windows.onFocusChanged.addListener((windowId) => {
  void (async () => {
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
  })();
});

chrome.idle.onStateChanged.addListener((state) => {
  if (state === "idle" || state === "locked") {
    void closeCurrent();
  }
});

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "flush") {
    void (async () => {
      await closeCurrent();
      await flush();
    })();
  }
});

chrome.runtime.onMessage.addListener(
  (msg: ExtensionMessage, _sender, sendResponse: (response?: unknown) => void) => {
    switch (msg.type) {
      case "ai-copy": {
        void handleAiCopy(msg).then(() => {
          const r: AiCopyResponse = { ok: true };
          sendResponse(r);
        });
        return true;
      }
      case "flush-now": {
        void closeCurrent()
          .then(flush)
          .then(() => {
            const r: FlushNowResponse = { ok: true };
            sendResponse(r);
          });
        return true;
      }
      case "pending-count": {
        void loadState().then((s) => {
          const r: PendingCountResponse = { buffer: s.buffer.length, retry: s.retryQueue.length };
          sendResponse(r);
        });
        return true;
      }
      case "set-paused": {
        void mutate((s) => {
          s.paused = msg.paused;
          if (msg.paused) s.current = undefined;
        }).then(() => {
          const r: SetPausedResponse = { ok: true };
          sendResponse(r);
        });
        return true;
      }
      case "get-status": {
        void loadState().then((s) => {
          const r: GetStatusResponse = {
            buffer: s.buffer.length,
            retry: s.retryQueue.length,
            paused: s.paused,
            lastSuccessTs: s.lastSuccessTs,
          };
          sendResponse(r);
        });
        return true;
      }
      default:
        return false;
    }
  },
);

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
  if (state.paused) return;
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

async function handleAiCopy(msg: { host: string; size: number }): Promise<void> {
  if (!msg.host) return;
  const info = aiInfoForHost(msg.host);
  if (!info) return;
  const state = await loadState();
  if (state.paused) return;
  await mutate((s) => {
    s.buffer.push({
      app_bundle: msg.host,
      category: "ai",
      source: "browser",
      ai_provider: info.provider,
      ai_channel: info.channel,
      duration_ms: 0,
      chars_in: msg.size,
    });
  });
}

async function flush(): Promise<void> {
  const state = await loadState();
  if (state.paused) return;
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
        s.lastSuccessTs = Date.now();
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
      void chrome.alarms.create("retry-flush", { delayInMinutes: delayMin });
      console.log(`[eop] retry-later (#${attempts}), next flush в ${delayMin}m`);
    }
  }
}

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "retry-flush") {
    void flush();
  }
});
