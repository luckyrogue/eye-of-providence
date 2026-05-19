import { useMemo } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";
import { useContent, type FaqItemsBlock, type FaqItem } from "@/shared/content";

// Auto-derive a kebab-case anchor from the question if the CMS row omits
// an explicit `anchor`. Strictly lowercase a-z 0-9 with single hyphens.
function slugifyQuestion(q: string): string {
  return q
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60);
}

export function FAQ() {
  const { t } = useTranslation("landing");

  // i18n is the synchronous fallback when the CMS row is missing/offline.
  // FAQ schema requires minItems: 3, so the i18n JSON ships at least 3
  // items per locale (Phase 4 i18n seed). Defensive .slice(0,10) for the
  // upper bound matches the backend's maxItems: 10.
  const fallback = useMemo<FaqItemsBlock>(() => {
    const items = t("faq.items", { returnObjects: true }) as FaqItem[];
    return { items: Array.isArray(items) ? items.slice(0, 10) : [] };
  }, [t]);

  const data = useContent<FaqItemsBlock>("landing.faq.items", fallback);
  const items = data.items;

  return (
    <section id="faq" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-3xl px-4 sm:px-6">
        <div className="mb-12">
          <Eyebrow>{t("faq.eyebrow")}</Eyebrow>
          <h2 className="display-head text-3xl sm:text-4xl md:text-5xl mt-3">
            <Trans i18nKey="landing:faq.heading" components={{ em: <em /> }} />
          </h2>
        </div>
        <div className="space-y-3">
          {items.map((it) => {
            const anchor = it.anchor ?? slugifyQuestion(it.q);
            return (
              <details key={it.q} id={`faq-${anchor}`} className="rounded-xl border bg-card group">
                <summary className="cursor-pointer list-none p-5 flex items-start justify-between gap-4 font-medium">
                  <span>{it.q}</span>
                  <span className="font-mono text-xl text-muted-foreground transition-transform group-open:rotate-45">
                    +
                  </span>
                </summary>
                <div className="px-5 pb-5 text-sm text-muted-foreground leading-relaxed">
                  {it.a}
                </div>
              </details>
            );
          })}
        </div>
      </div>
    </section>
  );
}
