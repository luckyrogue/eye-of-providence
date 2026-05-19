import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button, Card, CardContent, Eyebrow, Skeleton } from "@eop/ui";
import { Markdown } from "@/shared/lib/markdown";

export function MarkdownDocPage({
  title,
  markdownPath,
  backHref = "/",
  backLabel,
}: {
  title: string;
  markdownPath: string;
  backHref?: string;
  backLabel?: string;
}) {
  const { t } = useTranslation("landing");
  const [source, setSource] = useState<string | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    setSource(null);
    setError(false);
    fetch(markdownPath)
      .then((r) => (r.ok ? r.text() : Promise.reject(r.status)))
      .then(setSource)
      .catch(() => setError(true));
  }, [markdownPath]);

  const back = backLabel ?? t("docs.back_home");

  return (
    <main className="relative min-h-screen pt-[68px]">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[320px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-3xl px-4 sm:px-8 py-12 pb-20">
        <div className="mb-8">
          <Link to={backHref}>
            <Button variant="ghost" size="sm" className="mb-4 -ml-2">
              ← {back}
            </Button>
          </Link>
          <Eyebrow>{t("docs.eyebrow")}</Eyebrow>
          <h1 className="display-head text-3xl sm:text-4xl mt-3 font-display font-medium tracking-tight">
            {title}
          </h1>
        </div>

        {error ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">{t("docs.error")}</p>
            </CardContent>
          </Card>
        ) : source == null ? (
          <div className="space-y-3">
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-5/6" />
          </div>
        ) : (
          <article
            className="rounded-xl border p-6 sm:p-8"
            style={{ borderColor: "hsl(var(--border))" }}
          >
            <Markdown source={source} />
          </article>
        )}
      </div>
    </main>
  );
}
