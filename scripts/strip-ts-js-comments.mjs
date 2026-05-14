/**
 * Strip // and /* *\/ comments from TS/JS sources using the TypeScript printer.
 * Preserves triple-slash reference directives. Run from repo root:
 *   node scripts/strip-ts-js-comments.mjs
 */
import { createRequire } from "node:module";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "..");
const require = createRequire(path.join(root, "dashboard", "package.json"));
const ts = require("typescript");

const exts = new Set([".ts", ".tsx", ".mts", ".cts", ".js", ".jsx"]);
const skipDir = new Set([
  "node_modules",
  ".git",
  "dist",
  "build",
  ".output",
  "coverage",
  "playwright-report",
  ".turbo",
  "target",
  ".cursor",
]);

function scriptKind(filePath) {
  const ext = path.extname(filePath);
  if (ext === ".tsx" || ext === ".jsx") return ts.ScriptKind.TSX;
  if (ext === ".ts" || ext === ".mts" || ext === ".cts") return ts.ScriptKind.TS;
  return ts.ScriptKind.JS;
}

function stripFile(absPath) {
  const text = fs.readFileSync(absPath, "utf8");
  const kind = scriptKind(absPath);
  const sf = ts.createSourceFile(absPath, text, ts.ScriptTarget.Latest, true, kind);
  const printer = ts.createPrinter({
    newLine: ts.NewLineKind.LineFeed,
    removeComments: true,
  });
  const out = printer.printFile(sf);
  if (out === text) return false;
  fs.writeFileSync(absPath, out, "utf8");
  return true;
}

let changed = 0;
let failed = 0;

function walk(dir) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const name = ent.name;
    const full = path.join(dir, name);
    if (ent.isDirectory()) {
      if (skipDir.has(name)) continue;
      walk(full);
      continue;
    }
    if (!exts.has(path.extname(name))) continue;
    if (name.endsWith(".min.js")) continue;
    try {
      if (stripFile(full)) changed++;
    } catch (e) {
      failed++;
      console.error("FAIL", full, e.message);
    }
  }
}

walk(root);
console.log(`TS/JS strip: ${changed} files updated, ${failed} parse failures`);
