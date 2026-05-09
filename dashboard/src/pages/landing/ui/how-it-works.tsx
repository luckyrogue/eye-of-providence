import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";

type HowItem = { title: string; desc: string };

export function HowItWorks() {
  const { t } = useTranslation("landing");
  const items = t("how.items", { returnObjects: true }) as HowItem[];
  return (
    <section id="how" className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="max-w-2xl mb-16">
          <Eyebrow>{t("how.eyebrow")}</Eyebrow>
          <h2 className="display-head text-4xl md:text-5xl mt-3">
            <Trans i18nKey="landing:how.heading" components={{ em: <em /> }} />
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {items.map((s, i) => (
            <div key={s.title} className="rounded-xl border bg-card p-6">
              <div className="font-mono text-xs uppercase tracking-widest3 text-muted-foreground">
                {t("how.step_label", { num: String(i + 1).padStart(2, "0") })}
              </div>
              <h3 className="font-display font-bold tracking-tight text-lg mt-3">{s.title}</h3>
              <p className="text-sm text-muted-foreground mt-2 leading-relaxed">{s.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
