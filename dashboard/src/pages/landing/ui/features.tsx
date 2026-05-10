import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";
import { Activity, Brain, Code2, Lock, Sparkles, Zap } from "lucide-react";

// Иконки в коде, тексты — в локалях. Порядок сохраняется ↔ items[i] в JSON.
const ICONS = [Brain, Activity, Code2, Lock, Sparkles, Zap];

type FeatureItem = { title: string; desc: string };

export function Features() {
  const { t } = useTranslation("landing");
  const items = t("features.items", { returnObjects: true }) as FeatureItem[];
  return (
    <section id="features" className="py-24">
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <div className="max-w-2xl mb-16">
          <Eyebrow>{t("features.eyebrow")}</Eyebrow>
          <h2 className="display-head text-3xl sm:text-4xl md:text-5xl mt-3">
            <Trans i18nKey="landing:features.heading" components={{ em: <em /> }} />
          </h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {items.map(({ title, desc }, i) => {
            const Icon = ICONS[i] ?? Brain;
            return (
              <div key={title} className="rounded-xl border bg-card p-6 card-hover">
                <div className="h-10 w-10 rounded-lg bg-secondary flex items-center justify-center mb-4">
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="font-display font-bold tracking-tight text-lg">{title}</h3>
                <p className="text-sm text-muted-foreground mt-2 leading-relaxed">{desc}</p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
