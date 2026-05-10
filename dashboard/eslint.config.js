// Минимальный ESLint flat config — только guard-правила против реинтродукции
// inline-дубликатов, которые мы только что вычистили в "UI Dedup + Shadcn Audit".
// Не делаем полноценный stylistic linting (для этого есть tsc + Prettier).
import tseslint from "typescript-eslint";

export default tseslint.config(
  {
    ignores: ["dist/**", "node_modules/**", "public/**"],
  },
  {
    files: ["src/**/*.{ts,tsx}"],
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
          // Голый <button> запрещён — есть Button (shadcn) и IconButton (наш над shadcn).
          // Подсветим разработчика, если он опять начнёт городить inline tab-switcher
          // или dismiss-X через native <button>.
          selector: "JSXOpeningElement[name.name='button']",
          message:
            "Use <Button> (CTA) or <IconButton> (icon-only) from @eop/ui. For tab-switchers use <Tabs>.",
        },
        {
          // <input readOnly value={url} /> с Copy-кнопкой — это паттерн SecretField.
          // Не дублируем через бару input + Copy.
          selector: "JSXOpeningElement[name.name='input'] > JSXAttribute[name.name='readOnly']",
          message: "Use <SecretField> from @eop/ui for read-only-with-copy pattern.",
        },
      ],
    },
  },
);
