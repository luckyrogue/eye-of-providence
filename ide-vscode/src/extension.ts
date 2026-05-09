import * as vscode from "vscode";

// Attribution v1 в IDE:
// - Каждое изменение документа — onDidChangeTextDocument с массивом contentChanges.
// - Малые insert (< pasteThreshold chars и без replace) → typed.
// - Большие insert (>= pasteThreshold) или одномоментная вставка → ai_inline (proxy для Copilot/Cursor accept).
// - Удаление + большой insert (replace) → refactor.
// Данные накапливаются per-language и шлются батчем каждые flushInterval секунд.

type Bucket = {
  lang: string;
  category: "manual" | "ai" | "refactor" | "other";
  ai_channel?: "inline";
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

function onChange(e: vscode.TextDocumentChangeEvent) {
  if (e.document.uri.scheme !== "file") return;
  const lang = e.document.languageId;
  const threshold = getPasteThreshold();

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
    } else if (inserted >= threshold) {
      addToBucket(lang, "ai", {
        chars_in: inserted,
        lines_added: linesAdded,
        lines_removed: linesRemoved,
        ai_channel: "inline",
      });
    } else {
      addToBucket(lang, "manual", {
        chars_in: inserted,
        lines_added: linesAdded,
        lines_removed: linesRemoved,
      });
    }
  }
}

function bucketKey(lang: string, category: Bucket["category"], aiChannel?: string): string {
  return `${lang}::${category}::${aiChannel ?? ""}`;
}

function addToBucket(
  lang: string,
  category: Bucket["category"],
  patch: Partial<Bucket> & { ai_channel?: "inline" },
) {
  const key = bucketKey(lang, category, patch.ai_channel);
  let b = buckets.get(key);
  if (!b) {
    b = {
      lang,
      category,
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
  if (buckets.size === 0) {
    if (verbose) vscode.window.showInformationMessage("Eye of Providence: nothing to flush");
    return;
  }
  const events: EventPayload[] = Array.from(buckets.values()).map((b) => ({
    app_bundle: "com.microsoft.VSCode",
    category: b.category,
    source: "ide",
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
