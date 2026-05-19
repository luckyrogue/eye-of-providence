import path from "node:path";
import { fileURLToPath } from "node:url";
import { baseConfig } from "@eop/ui/eslint.base.js";

const dashboardRoot = path.dirname(fileURLToPath(import.meta.url));

export default baseConfig({
  fsdLayers: true,
  fsdSrcRoot: path.join(dashboardRoot, "src"),
});
