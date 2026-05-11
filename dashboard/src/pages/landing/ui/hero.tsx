import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";
import { ArrowRight, Check } from "lucide-react";
import { ProductPreview } from "./product-preview";

export function Hero() {
  const { t } = useTranslation("landing");
  const trust = t("hero.trust", { returnObjects: true }) as string[];
  return (
    <section className="relative pt-[140px] pb-20 sm:pb-28 px-4 sm:px-8">
      <div className="relative mx-auto max-w-7xl">
        <div className="text-center max-w-4xl mx-auto">
          <Eyebrow className="reveal">{t("hero.badge")}</Eyebrow>
          <h1 className="display-head text-[clamp(2.5rem,6vw,5.4rem)] mt-5 reveal reveal-delay-1 text-balance">
            <Trans i18nKey="landing:hero.heading" components={{ em: <em /> }} />
          </h1>
          <p className="mt-6 text-[clamp(0.95rem,1.4vw,1.125rem)] text-muted-foreground max-w-[58ch] mx-auto reveal reveal-delay-2 text-pretty">
            {t("hero.lead")}
          </p>
          <div className="mt-8 flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-3 reveal reveal-delay-3">
            <a
              href="/dashboard"
              className="btn-eop-primary inline-flex items-center justify-center gap-2 px-5 py-3 rounded-lg text-[15px] font-medium"
            >
              {t("hero.cta_primary")}
              <ArrowRight className="h-4 w-4" />
            </a>
            <a
              href="#how"
              className="inline-flex items-center justify-center px-5 py-3 rounded-lg text-[15px] font-medium border transition-colors hover:bg-foreground/5"
              style={{ borderColor: "hsl(var(--eop-line-strong))" }}
            >
              {t("hero.cta_secondary")}
            </a>
          </div>
          <div className="mt-10 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-[12px] font-mono text-muted-foreground reveal reveal-delay-4">
            {trust.map((s) => (
              <span key={s} className="inline-flex items-center gap-1.5">
                <Check className="h-3.5 w-3.5 text-[hsl(var(--accent))]" /> {s}
              </span>
            ))}
          </div>
        </div>

        <div className="mt-16 sm:mt-20 reveal reveal-delay-5">
          <ProductPreview />
        </div>
      </div>
    </section>
  );
}
