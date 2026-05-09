import { Trans, useTranslation } from "react-i18next";
import { Button } from "@eop/ui";
import { ArrowRight } from "lucide-react";

export function CTASection() {
  const { t } = useTranslation("landing");
  return (
    <section className="py-24">
      <div className="mx-auto max-w-3xl px-6 text-center">
        <h2 className="display-head text-5xl md:text-6xl">
          <Trans i18nKey="landing:cta.heading" components={{ em: <em /> }} />
        </h2>
        <p className="mt-5 text-muted-foreground max-w-md mx-auto">{t("cta.lead")}</p>
        <Button asChild size="lg" className="mt-8 h-12 px-8">
          <a href="/dashboard" className="flex items-center gap-2">
            {t("cta.button")}
            <ArrowRight className="h-4 w-4" />
          </a>
        </Button>
      </div>
    </section>
  );
}
