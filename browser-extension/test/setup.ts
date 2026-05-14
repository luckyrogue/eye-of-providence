type Listener = (changes: Record<string, chrome.storage.StorageChange>, areaName: string) => void;
const localStore = new Map<string, unknown>();
const onChangedListeners = new Set<Listener>();
function emit(changes: Record<string, chrome.storage.StorageChange>) {
  for (const fn of onChangedListeners) fn(changes, "local");
}
const storageLocal: Partial<chrome.storage.LocalStorageArea> = {
  async get(keys?: string | string[] | Record<string, unknown> | null) {
    const out: Record<string, unknown> = {};
    if (keys == null) {
      for (const [k, v] of localStore) out[k] = v;
    } else if (typeof keys === "string") {
      if (localStore.has(keys)) out[keys] = localStore.get(keys);
    } else if (Array.isArray(keys)) {
      for (const k of keys) if (localStore.has(k)) out[k] = localStore.get(k);
    } else {
      for (const k of Object.keys(keys)) {
        out[k] = localStore.has(k) ? localStore.get(k) : keys[k];
      }
    }
    return out;
  },
  async set(items: Record<string, unknown>) {
    const changes: Record<string, chrome.storage.StorageChange> = {};
    for (const [k, v] of Object.entries(items)) {
      changes[k] = { oldValue: localStore.get(k), newValue: v };
      localStore.set(k, v);
    }
    emit(changes);
  },
  async remove(keys: string | string[]) {
    const arr = Array.isArray(keys) ? keys : [keys];
    const changes: Record<string, chrome.storage.StorageChange> = {};
    for (const k of arr) {
      changes[k] = { oldValue: localStore.get(k), newValue: undefined };
      localStore.delete(k);
    }
    emit(changes);
  },
  async clear() {
    localStore.clear();
  },
};
const chromeMock = {
  storage: {
    local: storageLocal,
    onChanged: {
      addListener: (fn: Listener) => onChangedListeners.add(fn),
      removeListener: (fn: Listener) => onChangedListeners.delete(fn),
    },
  },
  runtime: {
    getURL: (path: string) => `chrome-extension://test/${path}`,
    onInstalled: { addListener: () => {} },
    onStartup: { addListener: () => {} },
    onMessage: { addListener: () => {} },
    sendMessage: async () => undefined,
  },
  tabs: { create: async () => undefined, query: async () => [] },
  alarms: {
    create: () => {},
    onAlarm: { addListener: () => {} },
  },
  idle: {
    setDetectionInterval: () => {},
    onStateChanged: { addListener: () => {} },
  },
  windows: { WINDOW_ID_NONE: -1, onFocusChanged: { addListener: () => {} } },
};
(
  globalThis as unknown as {
    chrome: typeof chromeMock;
  }
).chrome = chromeMock;
import { afterEach } from "vitest";
afterEach(() => {
  localStore.clear();
});
