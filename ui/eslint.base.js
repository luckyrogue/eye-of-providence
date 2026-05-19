import fs from "node:fs";
import path from "node:path";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";

export function baseConfig({
  files = ["src/**/*.{ts,tsx}"],
  extraRules = {},
  fsdLayers = false,
  fsdSrcRoot = null,
} = {}) {
  const blocks = [
    { ignores: ["dist/**", "node_modules/**", "public/**"] },
    {
      files,
      plugins: {
        "@typescript-eslint": tseslint.plugin,
        "react-hooks": reactHooks,
      },
      languageOptions: {
        parser: tseslint.parser,
        parserOptions: {
          ecmaFeatures: { jsx: true },
          sourceType: "module",
          projectService: true,
        },
      },
      rules: {
        "react-hooks/rules-of-hooks": "error",
        "react-hooks/exhaustive-deps": "warn",

        "@typescript-eslint/consistent-type-imports": [
          "warn",
          { prefer: "type-imports", fixStyle: "inline-type-imports" },
        ],
        "@typescript-eslint/no-floating-promises": "error",
        "@typescript-eslint/no-misused-promises": [
          "error",
          { checksVoidReturn: { attributes: false } },
        ],

        "no-restricted-syntax": [
          "error",
          {
            selector: "JSXOpeningElement[name.name='input']",
            message:
              "Use <Input> from @eop/ui (or <InputField> with react-hook-form). For read-only secrets use <SecretField>.",
          },
          {
            selector: "JSXOpeningElement[name.name='select']",
            message: "Use <Select> / <SimpleSelect> / <SelectField> from @eop/ui.",
          },
          {
            selector: "JSXOpeningElement[name.name='button']",
            message:
              "Use <Button> (CTA) or <IconButton> (icon-only) from @eop/ui. For tab-switchers use <Tabs>.",
          },
        ],
        ...extraRules,
      },
    },
  ];

  if (fsdLayers) {
    blocks.push(...fsdLayerBlocks());
    if (fsdSrcRoot) blocks.push(...fsdHorizontalBlocks(fsdSrcRoot));
    blocks.push(...fsdPublicApiBlocks());
  }
  return tseslint.config(...blocks);
}

const FSD_LAYERS = ["entities", "features", "widgets", "pages"];
const FSD_UPPER_BY_LAYER = {
  shared: ["entities", "features", "widgets", "pages", "app"],
  entities: ["features", "widgets", "pages", "app"],
  features: ["widgets", "pages", "app"],
  widgets: ["pages", "app"],
  pages: ["app"],
};

function verticalPatterns(layer) {
  const forbidden = FSD_UPPER_BY_LAYER[layer] ?? [];
  return forbidden.flatMap((upper) => [
    {
      group: [`**/${upper}/**`],
      message: `FSD: src/${layer} cannot import from src/${upper} (upper layer).`,
    },
    {
      group: [`@/${upper}/**`, `~/${upper}/**`],
      message: `FSD: src/${layer} cannot import from src/${upper} (upper layer).`,
    },
  ]);
}

function horizontalPatterns(layer, otherSlice) {
  const rel = (prefix) => [`${prefix}${otherSlice}`, `${prefix}${otherSlice}/**`];
  return [
    {
      group: [`@/${layer}/${otherSlice}`, `@/${layer}/${otherSlice}/**`],
      message: `FSD: cannot import slice ${layer}/${otherSlice} from another ${layer} slice.`,
    },
    {
      group: [`**/${layer}/${otherSlice}`, `**/${layer}/${otherSlice}/**`],
      message: `FSD: cannot import slice ${layer}/${otherSlice} from another ${layer} slice.`,
    },
    {
      group: [...rel("../"), ...rel("../../"), ...rel("../../../")],
      message: `FSD: cannot import slice ${layer}/${otherSlice} from another ${layer} slice (relative path).`,
    },
  ];
}

function listFsdSlices(srcRoot, layer) {
  const dir = path.join(srcRoot, layer);
  if (!fs.existsSync(dir)) return [];
  return fs
    .readdirSync(dir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && !entry.name.startsWith("."))
    .map((entry) => entry.name);
}

function fsdLayerBlocks() {
  const restrictFor = (layer, forbidden) => ({
    files: [`src/${layer}/**/*.{ts,tsx}`],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: forbidden.flatMap((upper) => [
            {
              group: [`**/${upper}/**`],
              message: `FSD: src/${layer} cannot import from src/${upper} (upper layer).`,
            },
            {
              group: [`@/${upper}/**`, `~/${upper}/**`],
              message: `FSD: src/${layer} cannot import from src/${upper} (upper layer).`,
            },
          ]),
        },
      ],
    },
  });
  return [restrictFor("shared", FSD_UPPER_BY_LAYER.shared)];
}

function fsdHorizontalBlocks(srcRoot) {
  const blocks = [];
  for (const layer of FSD_LAYERS) {
    const slices = listFsdSlices(srcRoot, layer);
    for (const slice of slices) {
      const others = slices.filter((name) => name !== slice);
      if (others.length === 0) continue;
      blocks.push({
        files: [`src/${layer}/${slice}/**/*.{ts,tsx}`],
        rules: {
          "no-restricted-imports": [
            "error",
            {
              patterns: [
                ...verticalPatterns(layer),
                ...others.flatMap((other) => horizontalPatterns(layer, other)),
              ],
            },
          ],
        },
      });
    }
  }
  return blocks;
}

function fsdPublicApiBlocks() {
  const entityInternals = [
    "@/entities/*/api/**",
    "@/entities/*/lib/**",
    "@/entities/*/model/**",
    "**/entities/*/api/**",
    "**/entities/*/lib/**",
    "**/entities/*/model/**",
  ];
  const featureInternals = [
    "@/features/*/ui/**",
    "@/features/*/api/**",
    "@/features/*/model/**",
    "@/features/*/lib/**",
    "**/features/*/ui/**",
    "**/features/*/api/**",
    "**/features/*/model/**",
    "**/features/*/lib/**",
  ];
  const widgetInternals = [
    "@/widgets/*/ui/**",
    "@/widgets/*/lib/**",
    "@/widgets/*/model/**",
    "**/widgets/*/ui/**",
    "**/widgets/*/lib/**",
    "**/widgets/*/model/**",
  ];
  const warnPatterns = (patterns, target) =>
    patterns.map((group) => ({
      group: [group],
      message: `FSD public API: import ${target} via its slice barrel (index), not internal segments.`,
    }));

  const consumerLayers = ["widgets", "features", "pages", "app"];
  return consumerLayers.map((layer) => ({
    files: [`src/${layer}/**/*.{ts,tsx}`],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            ...warnPatterns(entityInternals, "entities"),
            ...warnPatterns(featureInternals, "features"),
            ...warnPatterns(widgetInternals, "widgets"),
            {
              group: ["@/shared/api/**", "@/shared/config/**", "@/shared/hooks/**"],
              message:
                "FSD public API: import shared segments via @/shared/api|config|hooks, not deep paths.",
            },
          ],
        },
      ],
    },
  }));
}
