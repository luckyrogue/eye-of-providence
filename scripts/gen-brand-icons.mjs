#!/usr/bin/env node
/**
 * Регенерация PNG-иконок из ui/src/shared/assets/eop-app-icon.svg
 * (dashboard PWA, browser-extension, agent source, ide-vscode).
 *
 * Запуск: node scripts/gen-brand-icons.mjs
 * Требует: pnpm install (sharp в @eop/browser-extension).
 */

import { readFile, mkdir, copyFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(__dirname, "..");
const require = createRequire(resolve(ROOT, "browser-extension/package.json"));
const sharp = require("sharp");

const SRC = resolve(ROOT, "ui/src/shared/assets/eop-app-icon.svg");

async function writePng(svg, outPath, size) {
  await mkdir(dirname(outPath), { recursive: true });
  await sharp(svg, { density: 512 }).resize(size, size).png({ compressionLevel: 9 }).toFile(outPath);
  console.log(`wrote ${outPath}`);
}

async function main() {
  const svg = await readFile(SRC);

  for (const size of [16, 32, 48, 128]) {
    await writePng(
      svg,
      resolve(ROOT, `browser-extension/public/icons/icon-${size}.png`),
      size,
    );
  }

  for (const size of [192, 512]) {
    await writePng(svg, resolve(ROOT, `dashboard/public/icon-${size}.png`), size);
  }

  await writePng(svg, resolve(ROOT, "agent/src-tauri/icons/icon.png"), 1024);
  await writePng(svg, resolve(ROOT, "ide-vscode/icon.png"), 128);

  console.log("\nAgent: run `cd agent/src-tauri && pnpm tauri icon icons/icon.png` to refresh .icns/.ico");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
