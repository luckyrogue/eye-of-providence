import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: true,
    watch: {
      usePolling: true,
      interval: 300,
    },
  },
  build: {
    rollupOptions: {
      output: {
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
