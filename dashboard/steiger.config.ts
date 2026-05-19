import { defineConfig } from "steiger";
import fsd from "@feature-sliced/steiger-plugin";

export default defineConfig([
  ...fsd.configs.recommended,
  {
    ignores: ["**/test/**", "**/*.test.{ts,tsx}", "**/*.spec.{ts,tsx}"],
  },
  {
    rules: {
      "fsd/insignificant-slice": "off",
      "fsd/no-public-api-sidestep": "off",
      "fsd/no-segmentless-slices": "off",
      "fsd/excessive-slicing": "off",
      "fsd/segments-by-purpose": "off",
      "fsd/forbidden-imports": "error",
    },
  },
  {
    files: ["./src/shared/**"],
    rules: {
      "fsd/public-api": "off",
    },
  },
]);
