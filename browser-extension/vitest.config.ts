import { defineConfig } from "vitest/config";

// Отдельный vitest config — без @crxjs/vite-plugin (тот пытается читать
// manifest и виснет в тестах). Тесты гоняются под jsdom: chrome.storage
// API мокается в setup-файле.
export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./test/setup.ts"],
    include: ["src/**/*.test.ts", "src/**/*.test.tsx"],
  },
});
