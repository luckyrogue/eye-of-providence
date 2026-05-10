// Source: dashboard/public/changelog.json (regenerated через `pnpm
// changelog:gen` или CI). Conventional-commit-парсер на backend
// (cmd/changelog-gen) фильтрует publicTypes (feat/fix/perf/refactor/docs).

export type ChangelogType = "feat" | "fix" | "perf" | "refactor" | "docs";

export type ChangelogEntry = {
  hash: string;
  date: string;
  type: ChangelogType;
  scope?: string;
  breaking?: boolean;
  summary: string;
};

export type ChangelogDoc = {
  generated_at: string;
  entries: ChangelogEntry[];
};
