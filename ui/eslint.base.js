import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
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
  if (fsdLayers) blocks.push(...fsdLayerBlocks());
  return tseslint.config(...blocks);
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
  return [
    restrictFor("shared", ["entities", "features", "widgets", "pages", "app"]),
    restrictFor("entities", ["features", "widgets", "pages", "app"]),
    restrictFor("features", ["widgets", "pages", "app"]),
    restrictFor("widgets", ["pages", "app"]),
    restrictFor("pages", ["app"]),
  ];
}
