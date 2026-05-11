// CTA — final card с conic-gradient halo + macOS / Windows download buttons.

import { useTranslation } from "react-i18next";

export function CTASection() {
  const { t } = useTranslation("landing");
  return (
    <section className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px]">
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
              animation: "eop-rotate-slow 20s linear infinite",
            }}
          />
          <div className="relative z-[2] flex flex-col items-center gap-7">
            <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] max-w-[700px] text-balance">
              {t("cta.title1")}
              <em>{t("cta.titleItal")}</em>
              {t("cta.title2")}
            </h2>
            <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
              {t("cta.sub")}
            </p>
            <div className="flex flex-wrap gap-3 justify-center">
              <a
                href="/dashboard"
                className="btn-eop-primary inline-flex items-center gap-2 px-[18px] py-[11px] rounded-lg text-[14px] font-medium"
              >
                <svg viewBox="0 0 24 24" fill="currentColor" className="h-[18px] w-[18px]">
                  <path d="M17.05 12.5c0-2.7 2.2-4 2.3-4-1.3-1.8-3.2-2-3.9-2-1.6-.2-3.2 1-4 1-.8 0-2.1-1-3.4-1-1.8 0-3.4 1-4.3 2.6-1.8 3.2-.5 7.9 1.3 10.5.9 1.3 1.9 2.7 3.3 2.6 1.3-.05 1.8-.85 3.4-.85s2 .85 3.4.85c1.4 0 2.3-1.3 3.2-2.6 1-1.5 1.4-3 1.4-3-.1 0-2.7-1.05-2.7-4.1zM14.6 4.7c.7-.85 1.2-2.05 1.05-3.25-1 0-2.3.7-3 1.55-.65.75-1.3 2-1.1 3.15 1.15.1 2.35-.6 3.05-1.45z" />
                </svg>
                {t("cta.primary")}
              </a>
              <a
                href="/dashboard"
                className="inline-flex items-center gap-2 px-[18px] py-[11px] rounded-lg text-[14px] font-medium border transition-colors hover:bg-foreground/5"
                style={{
                  borderColor: "hsl(var(--eop-line-strong))",
                  background: "rgba(255,255,255,0.02)",
                }}
              >
                <svg viewBox="0 0 24 24" fill="currentColor" className="h-[18px] w-[18px]">
                  <path d="M3 5.5 11 4.4v7.1H3V5.5zm0 13L11 19.6v-7H3v6zm9 1.3 9 1.2v-8.5h-9v7.3zm0-15.5v7.2h9V3l-9 1.3z" />
                </svg>
                {t("cta.secondary")}
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
