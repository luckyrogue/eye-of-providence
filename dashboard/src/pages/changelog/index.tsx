// Public changelog — marketing page.
//
// Source: dashboard/public/changelog.json (regenerated через `pnpm
// changelog:gen` или CI). Conventional-commit-парсер на backend (cmd/
// changelog-gen) фильтрует publicTypes (feat/fix/perf/refactor/docs).
//
// Не требует auth — accessible вне dashboard session.

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, Eyebrow, Skeleton } from "@eop/ui";
import { Sparkles, Wrench, Zap, BookOpen, Hammer } from "lucide-react";
import type { LucideIcon } from "lucide-react";

type Entry = {
  hash: string;
  date: string;
  type: "feat" | "fix" | "perf" | "refactor" | "docs";
  scope?: string;
  breaking?: boolean;
  summary: string;
};

type Doc = {
  generated_at: string;
  entries: Entry[];
};

const TYPE_ICON: Record<Entry["type"], { icon: LucideIcon; color: string }> = {
  feat: { icon: Sparkles, color: "text-purple-500" },
  fix: { icon: Wrench, color: "text-amber-500" },
  perf: { icon: Zap, color: "text-emerald-500" },
  refactor: { icon: Hammer, color: "text-blue-500" },
  docs: { icon: BookOpen, color: "text-muted-foreground" },
};

export function ChangelogRoute() {
  const { t } = useTranslation("changelog");
  const [doc, setDoc] = useState<Doc | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetch("/changelog.json")
      .then((r) => (r.ok ? r.json() : Promise.reject(r.status)))
      .then(setDoc)
      .catch(() => setError(true));
  }, []);

  return (
    <main className="relative min-h-screen">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-3xl px-6 pt-12 pb-20">
        <div className="text-center mb-10 reveal">
          <Eyebrow>{t("eyebrow")}</Eyebrow>
          <h1 className="display-head text-4xl md:text-5xl mt-3">
            <em>{t("title")}</em>
          </h1>
          <p className="mt-3 text-sm text-muted-foreground">{t("lead")}</p>
        </div>

        {error ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">{t("error")}</p>
            </CardContent>
          </Card>
        ) : !doc ? (
          <div className="space-y-3">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        ) : doc.entries.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">{t("empty")}</p>
            </CardContent>
          </Card>
        ) : (
          <Timeline entries={doc.entries} />
        )}
      </div>
    </main>
  );
}

function Timeline({ entries }: { entries: Entry[] }) {
  // Группировка по дате (YYYY-MM-DD) — в одном дне может быть много commit'ов.
  const groups = new Map<string, Entry[]>();
  for (const e of entries) {
    const arr = groups.get(e.date) ?? [];
    arr.push(e);
    groups.set(e.date, arr);
  }
  const dates = Array.from(groups.keys());
  return (
    <ol className="space-y-8">
      {dates.map((date) => (
        <li key={date}>
          <DateHeading date={date} />
          <ul className="space-y-2 mt-3">
            {groups.get(date)!.map((e) => <Row key={e.hash} entry={e} />)}
          </ul>
        </li>
      ))}
    </ol>
  );
}

function DateHeading({ date }: { date: string }) {
  // Format YYYY-MM-DD → "10 May 2026" в локали юзера.
  const d = new Date(date + "T00:00:00Z");
  const formatted = d.toLocaleDateString(undefined, {
    day: "numeric",
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
  return (
    <h3 className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground">
      {formatted}
    </h3>
  );
}

function Row({ entry }: { entry: Entry }) {
  const { t } = useTranslation("changelog");
  const cfg = TYPE_ICON[entry.type];
  const Icon = cfg.icon;
  return (
    <li className="flex items-start gap-3 rounded-md border bg-muted/20 p-3 transition-colors hover:bg-muted/40">
      <Icon className={`h-4 w-4 mt-0.5 shrink-0 ${cfg.color}`} />
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className="text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
            {t(`type.${entry.type}` as const)}
          </span>
          {entry.scope && (
            <span className="text-[11px] font-mono px-1.5 py-0.5 rounded bg-muted">
              {entry.scope}
            </span>
          )}
          {entry.breaking && (
            <span className="text-[10px] uppercase font-semibold tracking-wider px-1.5 py-0.5 rounded bg-red-500/10 text-red-600 dark:text-red-400">
              {t("breaking")}
            </span>
          )}
        </div>
        <p className="text-sm mt-0.5">{entry.summary}</p>
      </div>
      <code className="text-[10px] font-mono text-muted-foreground tabular-nums shrink-0">
        {entry.hash}
      </code>
    </li>
  );
}
