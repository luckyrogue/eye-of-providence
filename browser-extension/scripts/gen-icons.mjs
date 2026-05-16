#!/usr/bin/env node
// Генерация PNG-иконок 16/32/48/128 из source SVG для MV3 manifest.
// Запуск: pnpm -F @eop/browser-extension gen-icons.

import { readFile, mkdir } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const __dirname = dirname(fileURLToPath(import.meta.url));
const SRC = resolve(__dirname, "..", "public", "icons", "icon.svg");
const OUT_DIR = resolve(__dirname, "..", "public", "icons");
const SIZES = [16, 32, 48, 128];

async function main() {
  const svg = await readFile(SRC);
  await mkdir(OUT_DIR, { recursive: true });
  for (const size of SIZES) {
    const out = resolve(OUT_DIR, `icon-${size}.png`);
    await sharp(svg, { density: 512 }).resize(size, size).png({ compressionLevel: 9 }).toFile(out);
    console.log(`wrote ${out}`);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
