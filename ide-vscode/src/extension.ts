import * as fs from "fs";
import * as os from "os";
import * as path from "path";
import * as vscode from "vscode";

// Attribution v2 в IDE:
// - Каждое изменение документа — onDidChangeTextDocument с массивом contentChanges.
// - Малые insert (< pasteThreshold chars и без replace) → typed.
// - Большие insert (>= pasteThreshold) → ai_inline (Copilot/Cursor accept).
//   ai_provider определяется через vscode.env.appName: Cursor → "cursor",
//   иначе "copilot" (по умолчанию VS Code inline-completions = Copilot).
// - Удаление + большой insert (replace) → refactor.
// - Burst-detection: несколько contentChanges за <100ms → один ai_inline event.
// Данные накапливаются per-language × ai_provider и шлются батчем каждые
// flushInterval секунд.

type AIProvider = "copilot" | "cursor";
type AIChannel = "inline" | "agent";

type Bucket = {
  lang: string;
  category: "manual" | "ai" | "refactor" | "other";
  ai_provider?: AIProvider;
  ai_channel?: AIChannel;
  duration_ms: number;
  chars_in: number;
  lines_added: number;
  lines_removed: number;
};

type EventPayload = {
  app_bundle: string;
  category: Bucket["category"];
  source: "ide";
  ai_provider?: string;
  ai_channel?: string;
  file_lang?: string;
  duration_ms: number;
  chars_in: number;
  lines_added: number;
  lines_removed: number;
};

type StatusKind = "idle" | "sending" | "auth-required" | "paused";

// SECRET_TOKEN_KEY — ключ в SecretStorage, в нём живёт API-токен (не path в
// глобальном config). После активации перекладываем сюда старый `eop.token`
// если он есть.
const SECRET_TOKEN_KEY = "eop.api_token";
const SECRET_USER_KEY = "eop.user_id";

// Persisted state в globalState — переживает рестарт VS Code, потерь батча
// между сессиями нет. Размер cap'нут чтобы не утечь память при долгом offline.
const STATE_QUEUE_KEY = "eop.queue";
const STATE_PAUSED_KEY = "eop.paused";
const MAX_QUEUE_SIZE = 2000;
const MAX_RETRY_ATTEMPTS = 8; // exp backoff: 30s, 60s, 2m, 4m, 8m, 16m, 32m, 60m

function detectAIProvider(): AIProvider {
  const name = vscode.env.appName.toLowerCase();
  if (name.includes("cursor")) return "cursor";
  return "copilot";
}

let buckets = new Map<string, Bucket>();
let activeEditorStart: number | null = null;
let activeLang: string | null = null;
let logger: vscode.LogOutputChannel;
let statusBar: vscode.StatusBarItem;
let secretStorage: vscode.SecretStorage;
let globalState: vscode.Memento;
let authRequired = false;
let paused = false;
let retryAttempt = 0;
let retryTimer: NodeJS.Timeout | null = null;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  logger = vscode.window.createOutputChannel("Eye of Providence", { log: true });
  statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBar.command = "eop.openDashboard";
  statusBar.show();
  secretStorage = context.secrets;
  globalState = context.globalState;
  paused = globalState.get<boolean>(STATE_PAUSED_KEY, false);

  await migrateLegacyToken();
  await renderStatus();

  logger.info("activated");

  context.subscriptions.push(
    statusBar,
    vscode.workspace.onDidChangeTextDocument(onChange),
    vscode.window.onDidChangeActiveTextEditor(onActiveEditorChange),
    vscode.workspace.onDidSaveTextDocument(() => flushFocus()),
    vscode.commands.registerCommand("eop.pair", pairCmd),
    vscode.commands.registerCommand("eop.logout", logoutCmd),
    vscode.commands.registerCommand("eop.flush", () => flushAll(true)),
    vscode.commands.registerCommand("eop.openDashboard", openDashboardCmd),
    vscode.commands.registerCommand("eop.showLog", () => logger.show()),
    vscode.commands.registerCommand("eop.pause", () => setPaused(true)),
    vscode.commands.registerCommand("eop.resume", () => setPaused(false)),
  );

  const interval = setInterval(() => {
    flushFocus();
    void flushAll(false);
  }, getFlushIntervalMs());
  context.subscriptions.push({ dispose: () => clearInterval(interval) });

  if (vscode.window.activeTextEditor) {
    onActiveEditorChange(vscode.window.activeTextEditor);
  }
}

export function deactivate() {
  flushFocus();
  void flushAll(false);
}

function onActiveEditorChange(editor: vscode.TextEditor | undefined) {
  flushFocus();
  if (editor) {
    activeEditorStart = Date.now();
    activeLang = editor.document.languageId;
  } else {
    activeEditorStart = null;
    activeLang = null;
  }
}

function flushFocus() {
  if (activeEditorStart === null || !activeLang) return;
  const ms = Date.now() - activeEditorStart;
  if (ms >= 1000) {
    addToBucket(activeLang, "other", { duration_ms: ms });
  }
  activeEditorStart = Date.now();
}

type BurstState = {
  start: number;
  inserted: number;
  linesAdded: number;
};
const bursts = new Map<string, BurstState>();
const BURST_WINDOW_MS = 100;

function onChange(e: vscode.TextDocumentChangeEvent) {
  if (paused) return;
  if (e.document.uri.scheme !== "file") return;
  if (!vscode.window.state.focused) return;
  const lang = e.document.languageId;
  const threshold = getPasteThreshold();
  const provider = detectAIProvider();
  const docKey = e.document.uri.toString();
  const now = Date.now();

  for (const c of e.contentChanges) {
    const inserted = c.text.length;
    const replaced = c.rangeLength;
    const linesAdded = c.text.split("\n").length - 1;
    const linesRemoved = replaced > 0 ? Math.max(0, e.document.getText(c.range).split("\n").length - 1) : 0;

    if (inserted === 0 && replaced > 0) {
      addToBucket(lang, "manual", { chars_in: 0, lines_removed: linesRemoved });
      continue;
    }

    if (replaced > threshold && inserted >= replaced * 0.5) {
      addToBucket(lang, "refactor", { chars_in: inserted, lines_added: linesAdded, lines_removed: linesRemoved });
      bursts.delete(docKey);
    } else if (inserted >= threshold) {
      addToBucket(lang, "ai", {
        chars_in: inserted,
        lines_added: linesAdded,
        lines_removed: linesRemoved,
        ai_provider: provider,
        ai_channel: "inline",
      });
      bursts.delete(docKey);
    } else {
      const burst = bursts.get(docKey);
      if (burst && now - burst.start <= BURST_WINDOW_MS) {
        burst.inserted += inserted;
        burst.linesAdded += linesAdded;
        if (burst.inserted >= threshold) {
          addToBucket(lang, "ai", {
            chars_in: burst.inserted,
            lines_added: burst.linesAdded,
            ai_provider: provider,
            ai_channel: "inline",
          });
          bursts.delete(docKey);
        }
      } else {
        bursts.set(docKey, { start: now, inserted, linesAdded });
        addToBucket(lang, "manual", {
          chars_in: inserted,
          lines_added: linesAdded,
          lines_removed: linesRemoved,
        });
      }
    }
  }
}

function bucketKey(
  lang: string,
  category: Bucket["category"],
  aiProvider?: string,
  aiChannel?: string,
): string {
  return `${lang}::${category}::${aiProvider ?? ""}::${aiChannel ?? ""}`;
}

function addToBucket(
  lang: string,
  category: Bucket["category"],
  patch: Partial<Bucket> & { ai_provider?: AIProvider; ai_channel?: AIChannel },
) {
  const key = bucketKey(lang, category, patch.ai_provider, patch.ai_channel);
  let b = buckets.get(key);
  if (!b) {
    b = {
      lang,
      category,
      ai_provider: patch.ai_provider,
      ai_channel: patch.ai_channel,
      duration_ms: 0,
      chars_in: 0,
      lines_added: 0,
      lines_removed: 0,
    };
    buckets.set(key, b);
  }
  b.duration_ms += patch.duration_ms ?? 0;
  b.chars_in += patch.chars_in ?? 0;
  b.lines_added += patch.lines_added ?? 0;
  b.lines_removed += patch.lines_removed ?? 0;
}

function bucketsToEvents(): EventPayload[] {
  const provider = detectAIProvider();
  const appBundle = provider === "cursor" ? "com.todesktop.230313mzl4w4u92" : "com.microsoft.VSCode";
  return Array.from(buckets.values()).map((b) => ({
    app_bundle: appBundle,
    category: b.category,
    source: "ide",
    ai_provider: b.ai_provider,
    ai_channel: b.ai_channel,
    file_lang: b.lang,
    duration_ms: b.duration_ms,
    chars_in: b.chars_in,
    lines_added: b.lines_added,
    lines_removed: b.lines_removed,
  }));
}

function loadQueue(): EventPayload[] {
  return globalState.get<EventPayload[]>(STATE_QUEUE_KEY, []);
}

async function saveQueue(q: EventPayload[]): Promise<void> {
  const capped = q.length > MAX_QUEUE_SIZE ? q.slice(-MAX_QUEUE_SIZE) : q;
  await globalState.update(STATE_QUEUE_KEY, capped);
}

async function flushAll(verbose: boolean) {
  if (paused) {
    if (verbose) vscode.window.showInformationMessage("Eye of Providence: paused");
    return;
  }
  if (!verbose && !vscode.window.state.focused) {
    return;
  }
  const fresh = bucketsToEvents();
  buckets = new Map();
  const queued = loadQueue();
  const events = [...queued, ...fresh];
  if (events.length === 0) {
    if (verbose) vscode.window.showInformationMessage("Eye of Providence: nothing to flush");
    return;
  }

  // README §3.3 architecture: events → local Tauri agent (encrypted SQLite +
  // offline queue + retry) → backend. Если agent запущен на этой машине,
  // шлём ему — events окажутся в едином буфере + получат re-emission при
  // offline. Если local agent недоступен (не запущен / другая машина / token
  // file отсутствует), падаем на cloud-direct ниже.
  if (await trySendToLocalAgent(events, verbose)) {
    await saveQueue([]);
    authRequired = false;
    retryAttempt = 0;
    await renderStatus();
    return;
  }

  const url = getBackendUrl();
  const token = await secretStorage.get(SECRET_TOKEN_KEY);
  if (!token) {
    // Persist обратно — попробуем после pairing.
    await saveQueue(events);
    if (verbose) {
      vscode.window
        .showWarningMessage("Eye of Providence: not paired", "Pair editor")
        .then((c) => {
          if (c === "Pair editor") void pairCmd();
        });
    }
    return;
  }

  setStatus("sending");
  try {
    const res = await fetch(`${url}/v1/ingest`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ events }),
    });
    if (res.status === 401) {
      logger.warn("ingest 401 — auth required, marking token invalid");
      authRequired = true;
      await saveQueue([]);
      await renderStatus();
      if (verbose) {
        vscode.window
          .showErrorMessage("Eye of Providence: token revoked. Re-pair this editor.", "Pair editor")
          .then((c) => {
            if (c === "Pair editor") void pairCmd();
          });
      }
      return;
    }
    if (res.status === 429 || res.status >= 500) {
      await saveQueue(events);
      scheduleRetry();
      logger.warn(`ingest retryable status ${res.status}, queued ${events.length} events`);
      return;
    }
    if (!res.ok) {
      logger.error(`ingest dropped ${events.length} events: status ${res.status}`);
      await saveQueue([]);
      return;
    }
    const data = (await res.json()) as { accepted: number; rejected: number };
    logger.info(`ingest ok: accepted=${data.accepted} rejected=${data.rejected}`);
    if (verbose) vscode.window.showInformationMessage(`EoP: sent ${data.accepted} events`);
    authRequired = false;
    retryAttempt = 0;
    await saveQueue([]);
  } catch (err) {
    await saveQueue(events);
    scheduleRetry();
    logger.warn(`ingest network error: ${String(err)}, queued ${events.length} events`);
  } finally {
    await renderStatus();
  }
}

function scheduleRetry() {
  retryAttempt += 1;
  if (retryAttempt > MAX_RETRY_ATTEMPTS) {
    logger.error(`retry exhausted after ${MAX_RETRY_ATTEMPTS} attempts, dropping queue`);
    void saveQueue([]);
    retryAttempt = 0;
    return;
  }
  const delayMs = Math.min(30_000 * Math.pow(2, retryAttempt - 1), 60 * 60_000);
  if (retryTimer !== null) clearTimeout(retryTimer);
  retryTimer = setTimeout(() => {
    retryTimer = null;
    void flushAll(false);
  }, delayMs);
  logger.info(`scheduled retry #${retryAttempt} in ${Math.round(delayMs / 1000)}s`);
}

async function setPaused(next: boolean) {
  paused = next;
  await globalState.update(STATE_PAUSED_KEY, paused);
  await renderStatus();
  vscode.window.showInformationMessage(
    next ? "Eye of Providence: paused" : "Eye of Providence: resumed",
  );
}

async function pairCmd() {
  const url = getBackendUrl();
  let begin: { pair_id: string; secret: string; code: string; expires_in: number };
  try {
    const res = await fetch(`${url}/v1/devices/pair`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ kind: "ide" }),
    });
    if (!res.ok) {
      vscode.window.showErrorMessage(`pair failed: ${res.status}`);
      return;
    }
    begin = (await res.json()) as typeof begin;
  } catch (err) {
    vscode.window.showErrorMessage(`pair error: ${String(err)}`);
    return;
  }

  const dashboardHost = url.replace(/\/api\/?$/, "");
  vscode.env.openExternal(vscode.Uri.parse(`${dashboardHost}/settings`));
  vscode.env.clipboard.writeText(begin.code);
  const pickPromise = vscode.window.showInformationMessage(
    `EoP pairing code: ${begin.code} (copied). Enter it in dashboard → Settings → Connected devices.`,
    "Cancel",
  );

  const start = Date.now();
  const expiresMs = begin.expires_in * 1000;
  let cancelled = false;
  pickPromise.then((choice) => {
    if (choice === "Cancel") cancelled = true;
  });

  while (!cancelled && Date.now() - start < expiresMs) {
    await sleep(2_500);
    try {
      const res = await fetch(`${url}/v1/devices/poll`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pair_id: begin.pair_id, secret: begin.secret }),
      });
      if (!res.ok) {
        logger.debug(`poll ${res.status}`);
        continue;
      }
      const data = (await res.json()) as {
        status: "pending" | "claimed" | "expired";
        token?: string;
        user_id?: string;
      };
      if (data.status === "expired") {
        vscode.window.showWarningMessage("EoP pairing code expired. Try again.");
        return;
      }
      if (data.status === "claimed" && data.token && data.user_id) {
        await secretStorage.store(SECRET_TOKEN_KEY, data.token);
        await secretStorage.store(SECRET_USER_KEY, data.user_id);
        authRequired = false;
        await renderStatus();
        vscode.window.showInformationMessage(
          `Eye of Providence: paired as ${data.user_id.slice(0, 8)}…`,
        );
        return;
      }
    } catch (err) {
      logger.debug(`poll error: ${String(err)}`);
    }
  }
  if (!cancelled) {
    vscode.window.showWarningMessage("Eye of Providence: pairing timed out.");
  }
}

async function logoutCmd() {
  await secretStorage.delete(SECRET_TOKEN_KEY);
  await secretStorage.delete(SECRET_USER_KEY);
  authRequired = false;
  await renderStatus();
  vscode.window.showInformationMessage("Eye of Providence: signed out.");
}

function openDashboardCmd() {
  const url = getBackendUrl().replace(/\/api\/?$/, "");
  void vscode.env.openExternal(vscode.Uri.parse(url));
}

// migrateLegacyToken — старый `eop.token` config был plaintext в settings.json.
// Переносим в SecretStorage и затираем config-key, чтобы не было утечек через
// settings sync.
async function migrateLegacyToken() {
  const cfg = vscode.workspace.getConfiguration("eop");
  const legacy = cfg.inspect<string>("token")?.globalValue;
  if (legacy && legacy.length > 0) {
    await secretStorage.store(SECRET_TOKEN_KEY, legacy);
    await cfg.update("token", undefined, vscode.ConfigurationTarget.Global);
    logger?.info("migrated legacy eop.token → SecretStorage");
  }
}

async function renderStatus() {
  const token = await secretStorage.get(SECRET_TOKEN_KEY);
  if (!token) {
    setStatus("auth-required");
    return;
  }
  if (authRequired) {
    setStatus("auth-required");
    return;
  }
  if (paused) {
    setStatus("paused");
    return;
  }
  setStatus("idle");
}

function setStatus(kind: StatusKind) {
  switch (kind) {
    case "idle":
      statusBar.text = "$(eye) EoP idle";
      statusBar.tooltip = "Eye of Providence — active. Click to open dashboard.";
      statusBar.backgroundColor = undefined;
      statusBar.command = "eop.openDashboard";
      break;
    case "sending":
      statusBar.text = "$(sync~spin) EoP sending";
      statusBar.tooltip = "Flushing attribution events…";
      statusBar.backgroundColor = undefined;
      statusBar.command = "eop.openDashboard";
      break;
    case "auth-required":
      statusBar.text = "$(warning) EoP pair";
      statusBar.tooltip = "Eye of Providence — re-pair required. Click to start.";
      statusBar.backgroundColor = new vscode.ThemeColor("statusBarItem.warningBackground");
      statusBar.command = "eop.pair";
      break;
    case "paused":
      statusBar.text = "$(circle-slash) EoP paused";
      statusBar.tooltip = "Eye of Providence — tracking paused.";
      statusBar.backgroundColor = undefined;
      statusBar.command = "eop.openDashboard";
      break;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function getBackendUrl(): string {
  const cfg = vscode.workspace.getConfiguration("eop");
  return (cfg.get<string>("backendUrl") ?? "https://eop.rysdavletov.org/api").replace(/\/$/, "");
}

// --- Local agent fallback (Block #2/4) -------------------------------------
//
// VS Code extension по умолчанию отправляет события в локальный Tauri-agent
// на 127.0.0.1, который буферизует их в encrypted SQLite, делает retry и сам
// шлёт в cloud. Это:
//   - даёт offline-resilience (если cloud недоступен, agent держит queue)
//   - объединяет potok событий от IDE, browser extension, OS-watcher,
//     CLI hooks в один pipeline (вместо четырёх отдельных HTTP-клиентов)
//   - соответствует архитектуре README §3.3
//
// Если local agent недоступен (не запущен / другая машина / token file
// отсутствует) — `trySendToLocalAgent` возвращает false и `flushAll` падает
// на cloud-direct (legacy path, остаётся для backwards-compat).

function getLocalAgentEnabled(): boolean {
  const cfg = vscode.workspace.getConfiguration("eop");
  return cfg.get<boolean>("localAgent.enabled", true);
}

function getLocalAgentUrl(): string {
  const cfg = vscode.workspace.getConfiguration("eop");
  return (cfg.get<string>("localAgent.url") ?? "http://127.0.0.1:7373").replace(/\/$/, "");
}

// localAgentDataDir — путь, куда Tauri пишет `eop.local-token`. Идентификатор
// `com.eyeofprovidence.agent` синхронизирован с agent/src-tauri/tauri.conf.json
// (поле `identifier`). Если кто-то меняет bundle id в Tauri — НЕ ЗАБЫТЬ обновить
// здесь.
function localAgentDataDir(): string {
  const ident = "com.eyeofprovidence.agent";
  switch (process.platform) {
    case "darwin":
      return path.join(os.homedir(), "Library", "Application Support", ident);
    case "win32":
      return path.join(process.env.APPDATA ?? os.homedir(), ident);
    default:
      return path.join(
        process.env.XDG_DATA_HOME ?? path.join(os.homedir(), ".local", "share"),
        ident,
      );
  }
}

// Кэшируем токен — файл не меняется в рамках одной сессии extension. Если
// agent был перезапущен и сгенерил новый токен, первый flush после рестарта
// получит 401 от local agent и упадёт на cloud — на следующем flush кеш
// инвалидируется (мы сбрасываем его на ошибку).
let cachedLocalToken: string | null | undefined = undefined;

async function readLocalAgentToken(): Promise<string | null> {
  if (cachedLocalToken !== undefined) return cachedLocalToken;
  try {
    const p = path.join(localAgentDataDir(), "eop.local-token");
    const buf = await fs.promises.readFile(p, "utf8");
    cachedLocalToken = buf.trim();
    return cachedLocalToken;
  } catch {
    cachedLocalToken = null;
    return null;
  }
}

async function trySendToLocalAgent(events: EventPayload[], verbose: boolean): Promise<boolean> {
  if (!getLocalAgentEnabled()) return false;
  const token = await readLocalAgentToken();
  if (!token) {
    logger.debug("local-agent: no token file, skipping");
    return false;
  }
  const url = getLocalAgentUrl();

  // 1.5s timeout — local loopback should be <50ms. Если agent повис, не
  // хотим блокировать flush на десятки секунд: лучше быстро упасть и пойти
  // в cloud напрямую.
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), 1500);
  try {
    setStatus("sending");
    const res = await fetch(`${url}/v1/local/ingest`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ events }),
      signal: ctrl.signal,
    });
    if (res.status === 401) {
      // Токен ротировал — сбросим кэш, следующий вызов перечитает.
      logger.warn("local-agent: 401, invalidating token cache");
      cachedLocalToken = undefined;
      return false;
    }
    if (!res.ok) {
      logger.debug(`local-agent: non-ok ${res.status}, falling back to cloud`);
      return false;
    }
    logger.info(`local-agent: forwarded ${events.length} events to 127.0.0.1`);
    if (verbose) {
      vscode.window.showInformationMessage(`EoP: sent ${events.length} events (via local agent)`);
    }
    return true;
  } catch (err) {
    // Connection refused / timeout / DNS — agent не запущен или
    // недоступен. Молчаливый fall back, иначе на каждой flush IDE будет
    // ругаться WARN'ами пока пользователь не запустит desktop app.
    logger.debug(`local-agent: unreachable (${String(err)}), falling back to cloud`);
    return false;
  } finally {
    clearTimeout(timer);
  }
}
// ---------------------------------------------------------------------------

function getFlushIntervalMs(): number {
  const cfg = vscode.workspace.getConfiguration("eop");
  return Math.max(5, cfg.get<number>("flushIntervalSec") ?? 30) * 1000;
}

function getPasteThreshold(): number {
  const cfg = vscode.workspace.getConfiguration("eop");
  return Math.max(10, cfg.get<number>("pasteThresholdChars") ?? 80);
}
