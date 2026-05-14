import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import preset from "./tailwind.preset.js";
const here = dirname(fileURLToPath(import.meta.url));
export function uiTailwindConfig({ content = [] } = {}) {
  return {
    presets: [preset],
    content: [...content, resolve(here, "./src/**/*.{ts,tsx}")],
  };
}
