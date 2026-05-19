import { defineConfig } from "steiger";
import fsd from "@feature-sliced/steiger-plugin";

export default defineConfig([
  ...fsd.configs.recommended,
  {
    ignores: ["**/test/**", "**/*.test.{ts,tsx}", "**/*.spec.{ts,tsx}"],
  },
  {
    rules: {
      "fsd/insignificant-slice": "error",
      "fsd/no-public-api-sidestep": "error",
      "fsd/no-segmentless-slices": "error",
      "fsd/excessive-slicing": "warn",
      "fsd/segments-by-purpose": "off",
      "fsd/forbidden-imports": "error",
    },
  },
  // Widgets are page-scoped composition blocks; Steiger would merge each into a single page.
  {
    files: ["./src/widgets/**"],
    rules: {
      "fsd/insignificant-slice": "off",
    },
  },
  // Domain entities may have a single primary page while staying reusable data layers.
  {
    files: ["./src/entities/**"],
    rules: {
      "fsd/insignificant-slice": "off",
    },
  },
  {
    files: ["./src/shared/**"],
    rules: {
      "fsd/public-api": "error",
    },
  },
]);
