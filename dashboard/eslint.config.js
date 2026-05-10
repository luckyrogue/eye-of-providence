// Минимальный ESLint flat config — guard-правила против реинтродукции
// inline-дубликатов из "UI Dedup + Shadcn Audit" задаются в общей базе.
// Не делаем полноценный stylistic linting (для этого есть tsc + Prettier).
import { baseConfig } from "@eop/ui/eslint.base.js";

export default baseConfig();
