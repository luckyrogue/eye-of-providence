import tseslint from "typescript-eslint";

// Базовый flat ESLint config для всех frontend-пакетов.
// Каждый пакет может extend и добавить свои overrides (chrome globals и т.п.).
export function baseConfig({ files = ["src/**/*.{ts,tsx}"], extraRules = {} } = {}) {
  return tseslint.config(
    { ignores: ["dist/**", "node_modules/**", "public/**"] },
    {
      files,
      languageOptions: {
        parser: tseslint.parser,
        parserOptions: {
          ecmaFeatures: { jsx: true },
          sourceType: "module",
        },
      },
      rules: {
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
  );
}
