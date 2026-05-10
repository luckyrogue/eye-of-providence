import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: true,
    watch: {
      // Polling нужен в Docker volume mount на macOS
      usePolling: true,
      interval: 300,
    },
  },
  build: {
    rollupOptions: {
      output: {
        // Стабильное разбиение vendor-chunks: тяжёлые libs кэшируются независимо
        // от правок приложения. Используем функцию: Radix-пакеты приходят
        // транзитивно через @eop/ui и могут отсутствовать в dependencies.
        manualChunks(id) {
          if (!id.includes("node_modules")) return undefined;
          if (id.includes("/react-router")) return "react-vendor";
          if (id.match(/\/react(-dom)?\//)) return "react-vendor";
          if (id.includes("@tanstack/react-query")) return "query-vendor";
          if (id.includes("@tanstack/react-table")) return "table-vendor";
          if (id.includes("react-hook-form")) return "form-vendor";
          if (id.includes("@hookform/resolvers")) return "form-vendor";
          if (id.includes("/zod/")) return "form-vendor";
          if (id.includes("i18next")) return "i18n-vendor";
          if (id.includes("react-i18next")) return "i18n-vendor";
          if (id.includes("@radix-ui/")) return "ui-radix";
          if (id.includes("/sonner/")) return "ui-radix";
          if (id.includes("lucide-react")) return "ui-radix";
          return undefined;
        },
      },
    },
  },
});
