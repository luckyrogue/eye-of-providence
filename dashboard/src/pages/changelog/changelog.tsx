// Public changelog — marketing page. Не требует auth.

import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Card, CardContent, Eyebrow, Skeleton } from "@eop/ui";
import type { ChangelogDoc } from "./model/types";
import { Timeline } from "./ui/timeline";

export function Changelog() {
  const { t } = useTranslation("changelog");
  const [doc, setDoc] = useState<ChangelogDoc | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetch("/changelog.json")
      .then((r) => (r.ok ? r.json() : Promise.reject(r.status)))
      .then(setDoc)
      .catch(() => setError(true));
  }, []);

  return (
    <main className="relative min-h-screen pt-[68px]">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-3xl px-4 sm:px-6 pt-12 pb-20">
        <div className="text-center mb-10 reveal">
          <Eyebrow>{t("eyebrow")}</Eyebrow>
          <h1 className="display-head text-3xl sm:text-4xl md:text-5xl mt-3">
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
