import { Trans, useTranslation } from "react-i18next";
import { Button, Eyebrow } from "@eop/ui";
import { ArrowRight, Check } from "lucide-react";
import { ProductPreview } from "./product-preview";

export function Hero() {
  const { t } = useTranslation("landing");
  const trust = t("hero.trust", { returnObjects: true }) as string[];
  return (
    <section className="relative overflow-hidden">
      <div className="dot-grid pointer-events-none absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
      <div className="absolute -top-40 left-1/2 -translate-x-1/2 h-[600px] w-[1100px] bg-gradient-to-br from-purple-500/10 via-blue-500/5 to-transparent blur-3xl pointer-events-none" />

      <div className="relative mx-auto max-w-6xl px-4 sm:px-6 pt-12 pb-16 sm:pt-20 sm:pb-24 text-center">
        <Eyebrow className="reveal">{t("hero.badge")}</Eyebrow>
        <h1 className="display-head text-4xl sm:text-5xl md:text-6xl lg:text-7xl mt-4 sm:mt-5 max-w-3xl mx-auto reveal reveal-delay-1">
          <Trans i18nKey="landing:hero.heading" components={{ em: <em /> }} />
        </h1>
        <p className="mt-4 sm:mt-6 text-base sm:text-lg text-muted-foreground max-w-xl mx-auto reveal reveal-delay-2">
          {t("hero.lead")}
        </p>
        {/* CTAs: column-stack на mobile (full-width buttons), horizontal на sm+ */}
        <div className="mt-6 sm:mt-8 flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-2 sm:gap-3 reveal reveal-delay-3">
          <Button asChild size="lg" className="h-11 px-6">
            <a href="/dashboard" className="flex items-center justify-center gap-2">
              {t("hero.cta_primary")}
              <ArrowRight className="h-4 w-4" />
            </a>
          </Button>
          <Button asChild size="lg" variant="outline" className="h-11 px-6">
            <a href="#how" className="flex items-center justify-center">{t("hero.cta_secondary")}</a>
          </Button>
        </div>
        {/* Trust strip: wrap-able grid на mobile */}
        <div className="mt-8 sm:mt-12 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 sm:gap-6 text-xs text-muted-foreground font-mono reveal reveal-delay-4">
          {trust.map((s) => (
            <span key={s} className="flex items-center gap-1.5">
              <Check className="h-3.5 w-3.5 text-foreground" /> {s}
            </span>
          ))}
        </div>

        <ProductPreview />
      </div>
    </section>
  );
}
