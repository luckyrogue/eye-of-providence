import * as vscode from "vscode";

// Attribution v2 в IDE:
// - Каждое изменение документа — onDidChangeTextDocument с массивом contentChanges.
// - Малые insert (< pasteThreshold chars и без replace) → typed.
// - Большие insert (>= pasteThreshold) → ai_inline (Copilot/Cursor accept).
//   ai_provider определяется через vscode.env.appName: Cursor → "cursor",
//   иначе "copilot" (по умолчанию VS Code inline-completions = Copilot).
// - Удаление + большой insert (replace) → refactor.
// - Burst-detection: несколько contentChanges за <100ms → один ai_inline event
//   (это inline streaming completion vs cmd+v paste = одно contentChange).
// Данные накапливаются per-language × ai_provider и шлются батчем каждые flushInterval секунд.

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

// detectAIProvider — Cursor — VS Code fork с собственным AI; appName === "Cursor".
// Остальные derivative IDE'шки (VSCodium, Windsurf и т.п.) тоже могут наследовать
// inline-completion provider от Copilot — fallback на "copilot".
function detectAIProvider(): AIProvider {
  const name = vscode.env.appName.toLowerCase();
  if (name.includes("cursor")) return "cursor";
  return "copilot";
}

let buckets = new Map<string, Bucket>();
let activeEditorStart: number | null = null;
let activeLang: string | null = null;
let logger: vscode.OutputChannel;

export function activate(context: vscode.ExtensionContext) {
  logger = vscode.window.createOutputChannel("Eye of Providence");
  logger.appendLine("activated");

  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument(onChange),
    vscode.window.onDidChangeActiveTextEditor(onActiveEditorChange),
    vscode.workspace.onDidSaveTextDocument(() => flushFocus()),
    vscode.commands.registerCommand("eop.devLogin", devLoginCmd),
    vscode.commands.registerCommand("eop.flush", () => flushAll(true)),
  );

  // Periodic flush
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

// Burst-detection state. Inline streaming completion пишет несколько мелких
// contentChanges подряд за <100ms; обычный typing — interval >150ms между ними.
// Если >=N changes сошлись в одно burst-окно и сумма chars >= threshold, это
// ai_inline. Хранится per-document.
type BurstState = {
  start: number;
  inserted: number;
  linesAdded: number;
};
const bursts = new Map<string, BurstState>();
const BURST_WINDOW_MS = 100;

function onChange(e: vscode.TextDocumentChangeEvent) {
  if (e.document.uri.scheme !== "file") return;
  // Multi-window dedup: VS Code broadcastит onDidChangeTextDocument во все
  // окна, где открыт документ. Считает событие только окно, у которого OS-фокус
  // (изменения user'а пришли через keyboard в одном конкретном окне).
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
      // Single big insert — paste/accept, attributed to ai_inline.
      addToBucket(lang, "ai", {
        chars_in: inserted,
        lines_added: linesAdded,
        lines_removed: linesRemoved,
        ai_provider: provider,
        ai_channel: "inline",
      });
      bursts.delete(docKey);
    } else {
      // Маленькая вставка — может быть typing или часть inline-streaming burst'а.
      const burst = bursts.get(docKey);
      if (burst && now - burst.start <= BURST_WINDOW_MS) {
        burst.inserted += inserted;
        burst.linesAdded += linesAdded;
        if (burst.inserted >= threshold) {
          // Burst накопился до AI-уровня — отписываем как ai_inline и закрываем.
          addToBucket(lang, "ai", {
            chars_in: burst.inserted,
            lines_added: burst.linesAdded,
            ai_provider: provider,
            ai_channel: "inline",
          });
          bursts.delete(docKey);
        }
      } else {
        // Новый burst или typing — начинаем burst-window и одновременно
        // counting как manual (если burst не превысит threshold, останется
        // manual; если превысит — последний chunk пере-attribут'нется).
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

async function flushAll(verbose: boolean) {
  // Multi-window dedup для periodic flush: только focused окно делает
  // network-call. Если фокус потерян (например, юзер ушёл в браузер) —
  // ждём, пока фокус вернётся, потом всё ещё накопленное в buckets улетит.
  // Manual flush (verbose=true) от пользователя — не блокируем focus-check.
  if (!verbose && !vscode.window.state.focused) {
    return;
  }
  if (buckets.size === 0) {
    if (verbose) vscode.window.showInformationMessage("Eye of Providence: nothing to flush");
    return;
  }
  // app_bundle отражает реальный host (Cursor / VS Code), чтобы analytics
  // могли filter'нуть по IDE.
  const provider = detectAIProvider();
  const appBundle = provider === "cursor" ? "com.todesktop.230313mzl4w4u92" : "com.microsoft.VSCode";
  const events: EventPayload[] = Array.from(buckets.values()).map((b) => ({
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
  buckets = new Map();

  const cfg = vscode.workspace.getConfiguration("eop");
  const url = (cfg.get<string>("backendUrl") ?? "https://eop.rysdavletov.org/api").replace(/\/$/, "");
  const token = cfg.get<string>("token") ?? "";
  if (!token) {
    if (verbose) vscode.window.showWarningMessage("Eye of Providence: no token (run 'eop.devLogin')");
    return;
  }

  try {
    const res = await fetch(`${url}/v1/ingest`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ events }),
    });
    if (!res.ok) {
      logger.appendLine(`ingest failed: ${res.status}`);
    } else {
      const data = (await res.json()) as { accepted: number; rejected: number };
      logger.appendLine(`ingest ok: accepted=${data.accepted} rejected=${data.rejected}`);
      if (verbose) vscode.window.showInformationMessage(`EoP: sent ${data.accepted} events`);
    }
  } catch (err) {
    logger.appendLine(`ingest error: ${err}`);
  }
}

async function devLoginCmd() {
  const cfg = vscode.workspace.getConfiguration("eop");
  const url = (cfg.get<string>("backendUrl") ?? "https://eop.rysdavletov.org/api").replace(/\/$/, "");
  try {
    const res = await fetch(`${url}/v1/auth/dev-token`, { method: "POST" });
    if (!res.ok) {
      vscode.window.showErrorMessage(`dev-token failed: ${res.status}`);
      return;
    }
    const data = (await res.json()) as { token: string; user_id: string };
    await cfg.update("token", data.token, vscode.ConfigurationTarget.Global);
    vscode.window.showInformationMessage(`EoP: logged in as ${data.user_id.slice(0, 8)}…`);
  } catch (err) {
    vscode.window.showErrorMessage(`dev-token error: ${err}`);
  }
}

function getFlushIntervalMs(): number {
  const cfg = vscode.workspace.getConfiguration("eop");
  return Math.max(5, cfg.get<number>("flushIntervalSec") ?? 30) * 1000;
}

function getPasteThreshold(): number {
  const cfg = vscode.workspace.getConfiguration("eop");
  return Math.max(10, cfg.get<number>("pasteThresholdChars") ?? 80);
}
