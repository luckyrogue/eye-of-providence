import type { ReactNode } from "react";
import { Link, useParams } from "react-router-dom";
import { Button } from "@eop/ui";
import { useTranslation } from "react-i18next";
import { MarketingLayout } from "@/widgets/marketing-layout";
import { LEGAL_PAGES, PUBLIC_DOCS, type LegalSlug } from "./model/docs-manifest";
import { DocsIndex } from "./ui/docs-index";
import { MarkdownDocPage } from "./ui/markdown-doc-page";

function DocLayout({ children }: { children: ReactNode }) {
  return <MarketingLayout>{children}</MarketingLayout>;
}

export function DocsIndexRoute() {
  return (
    <DocLayout>
      <DocsIndex />
    </DocLayout>
  );
}

export function DocSlugRoute() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation("landing");
  const entry = PUBLIC_DOCS.find((d) => d.slug === slug);

  if (!entry) {
    return (
      <DocLayout>
        <main className="min-h-screen pt-[68px] flex items-center justify-center px-6">
          <div className="text-center space-y-4">
            <p className="text-sm text-muted-foreground">{t("docs.error")}</p>
            <Link to="/docs">
              <Button variant="outline">{t("docs.back_docs")}</Button>
            </Link>
          </div>
        </main>
      </DocLayout>
    );
  }

  return (
    <DocLayout>
      <MarkdownDocPage
        title={t(entry.titleKey)}
        markdownPath={entry.path}
        backHref="/docs"
        backLabel={t("docs.back_docs")}
      />
    </DocLayout>
  );
}

export function LegalRoute({ page }: { page: LegalSlug }) {
  const { t } = useTranslation("landing");
  const meta = LEGAL_PAGES[page];

  return (
    <DocLayout>
      <MarkdownDocPage title={t(meta.titleKey)} markdownPath={meta.path} />
    </DocLayout>
  );
}

export function PrivacyRoute() {
  return <LegalRoute page="privacy" />;
}

export function TermsRoute() {
  return <LegalRoute page="terms" />;
}

export function SecurityRoute() {
  return <LegalRoute page="security" />;
}
