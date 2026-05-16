// Базовый Tailwind config для всех frontend-пакетов, потребляющих @eop/ui.
// Принимает массив локальных content-paths. Сам пакет тащит content из ui/src/**.
//
// Использование (dashboard/tailwind.config.js):
//   import { uiTailwindConfig } from "@eop/ui/tailwind.config.js";
//   export default uiTailwindConfig({ content: ["./index.html", "./src/**/*.{ts,tsx}"] });
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import preset from "./tailwind.preset.js";

const here = dirname(fileURLToPath(import.meta.url));

/**
 * @param {{ content?: string[] }} [options]
 * @returns {import('tailwindcss').Config}
 */
export function uiTailwindConfig({ content = [] } = {}) {
  return {
    presets: [preset],
    content: [...content, resolve(here, "./src/**/*.{ts,tsx}")],
  };
}
