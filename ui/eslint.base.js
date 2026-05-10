import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";

// Базовый flat ESLint config для всех frontend-пакетов.
// Каждый пакет может extend и добавить свои overrides (chrome globals и т.п.).
//
// Опции:
// - `files` — glob, по умолчанию src/**/*.{ts,tsx}
// - `extraRules` — доп. правила/overrides
// - `fsdLayers` — true, если пакет следует Feature-Sliced Design (запрет cross-layer импортов
//   из верхних слоёв в нижние). Включаем точечно в `dashboard` (т.к. agent/extension/ui
//   используют упрощённый или иной набор слоёв).
export function baseConfig({
  files = ["src/**/*.{ts,tsx}"],
  extraRules = {},
  fsdLayers = false,
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
          // projectService: true включает type-aware правила без явного project-ссылки.
          // Поддерживается typescript-eslint >= 8.
          projectService: true,
        },
      },
      rules: {
        // ---- React hooks
        "react-hooks/rules-of-hooks": "error",
        "react-hooks/exhaustive-deps": "warn",

        // ---- TypeScript hygiene
        "@typescript-eslint/consistent-type-imports": [
          "warn",
          { prefer: "type-imports", fixStyle: "inline-type-imports" },
        ],
        "@typescript-eslint/no-floating-promises": "error",
        "@typescript-eslint/no-misused-promises": [
          "error",
          { checksVoidReturn: { attributes: false } },
        ],

        // ---- Anti-pattern guards (UI primitives)
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

  if (fsdLayers) blocks.push(...fsdLayerBlocks());
  return tseslint.config(...blocks);
}

// FSD cross-layer restrictions для пакетов с полной структурой
// (app, pages, widgets, features, entities, shared). Реализовано через
// `no-restricted-imports.patterns` с file-overrides.
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
  return [
    restrictFor("shared", ["entities", "features", "widgets", "pages", "app"]),
    restrictFor("entities", ["features", "widgets", "pages", "app"]),
    restrictFor("features", ["widgets", "pages", "app"]),
    restrictFor("widgets", ["pages", "app"]),
    restrictFor("pages", ["app"]),
  ];
}
