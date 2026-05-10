import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";

type FaqItem = { q: string; a: string };

export function FAQ() {
  const { t } = useTranslation("landing");
  const items = t("faq.items", { returnObjects: true }) as FaqItem[];
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
          {items.map((it) => (
            <details key={it.q} className="rounded-xl border bg-card group">
              <summary className="cursor-pointer list-none p-5 flex items-start justify-between gap-4 font-medium">
                <span>{it.q}</span>
                <span className="font-mono text-xl text-muted-foreground transition-transform group-open:rotate-45">+</span>
              </summary>
              <div className="px-5 pb-5 text-sm text-muted-foreground leading-relaxed">{it.a}</div>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
