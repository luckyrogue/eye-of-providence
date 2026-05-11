import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";

type HowItem = { title: string; desc: string };

export function HowItWorks() {
  const { t } = useTranslation("landing");
  const items = t("how.items", { returnObjects: true }) as HowItem[];
  return (
    <section id="how" className="py-24 sm:py-32 border-t px-4 sm:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="max-w-2xl mb-14 sm:mb-16">
          <Eyebrow>{t("how.eyebrow")}</Eyebrow>
          <h2 className="display-head text-[clamp(1.9rem,4vw,3.4rem)] mt-3 text-balance">
            <Trans i18nKey="landing:how.heading" components={{ em: <em /> }} />
          </h2>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 sm:gap-6">
          {items.map((s, i) => (
            <div
              key={s.title}
              className="p-6 rounded-xl border bg-card/40 min-h-[220px] flex flex-col gap-3.5"
              style={{ borderColor: "hsl(var(--border))" }}
            >
              <div className="font-mono text-[11px] tracking-widest3 text-[hsl(var(--accent))]">
                {String(i + 1).padStart(2, "0")}
              </div>
              <h3 className="font-sans font-medium tracking-tight text-[18px] leading-tight">
                {s.title}
              </h3>
              <p className="text-[13px] text-muted-foreground leading-relaxed">{s.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
