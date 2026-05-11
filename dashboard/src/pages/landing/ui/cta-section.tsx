import { Trans, useTranslation } from "react-i18next";
import { ArrowRight } from "lucide-react";

export function CTASection() {
  const { t } = useTranslation("landing");
  return (
    <section className="py-24 sm:py-32 px-4 sm:px-8">
      <div className="mx-auto max-w-5xl">
        {/* CTA card: dark gradient + conic-glow halo вокруг */}
        <div
          className="relative overflow-hidden rounded-3xl border px-8 py-16 sm:px-12 sm:py-20 text-center flex flex-col items-center gap-7"
          style={{
            borderColor: "hsl(var(--eop-line-strong))",
            background:
              "radial-gradient(ellipse 80% 100% at 50% 0%, hsl(var(--accent) / 0.18), transparent 70%), linear-gradient(180deg, hsl(var(--card)), hsl(var(--background)))",
          }}
        >
          <div
            className="absolute inset-[-1px] pointer-events-none opacity-30"
            aria-hidden
            style={{
              background:
                "conic-gradient(from 0deg, transparent, hsl(var(--accent)), transparent 30%)",
              filter: "blur(40px)",
              animation: "spin 20s linear infinite",
            }}
          />
          <div className="relative z-[2] flex flex-col items-center gap-7">
            <h2 className="display-head text-[clamp(2rem,4.5vw,3.75rem)] text-balance">
              <Trans i18nKey="landing:cta.heading" components={{ em: <em /> }} />
            </h2>
            <p className="text-muted-foreground max-w-md text-[15px]">{t("cta.lead")}</p>
            <a
              href="/dashboard"
              className="btn-eop-primary inline-flex items-center gap-2 px-6 py-3 rounded-lg text-[15px] font-medium"
            >
              {t("cta.button")}
              <ArrowRight className="h-4 w-4" />
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
