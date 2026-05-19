import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";
import { PUBLIC_DOCS } from "../model/docs-manifest";

export function DocsIndex() {
  const { t } = useTranslation("landing");

  return (
    <main className="relative min-h-screen pt-[68px]">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[320px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-3xl px-4 sm:px-8 py-12 pb-20">
        <div className="text-center mb-10">
          <Eyebrow>{t("docs.eyebrow")}</Eyebrow>
          <h1 className="display-head text-3xl sm:text-4xl mt-3 font-display font-medium tracking-tight">
            {t("docs.title")}
          </h1>
          <p className="mt-3 text-sm text-muted-foreground max-w-xl mx-auto">{t("docs.lead")}</p>
        </div>
        <ul className="space-y-3">
          {PUBLIC_DOCS.map((doc) => (
            <li key={doc.slug}>
              <Link
                to={`/docs/${doc.slug}`}
                className="block rounded-xl border px-5 py-4 transition-colors hover:bg-foreground/5"
                style={{ borderColor: "hsl(var(--border))" }}
              >
                <span className="font-medium">{t(doc.titleKey)}</span>
                <p className="text-sm text-muted-foreground mt-1">{t(`${doc.titleKey}_lead`)}</p>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
