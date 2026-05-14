import { Trans, useTranslation } from "react-i18next";
import { Eyebrow } from "@eop/ui";
import { Activity, Brain, Code2, Lock, Sparkles, Zap } from "lucide-react";
const ICONS = [Brain, Activity, Code2, Lock, Sparkles, Zap];
type FeatureItem = {
  title: string;
  desc: string;
};
export function Features() {
  const { t } = useTranslation("landing");
  const items = t("features.items", { returnObjects: true }) as FeatureItem[];
  return (
    <section id="features" className="py-24 sm:py-32 px-4 sm:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="max-w-2xl mb-14 sm:mb-16">
          <Eyebrow>{t("features.eyebrow")}</Eyebrow>
          <h2 className="display-head text-[clamp(1.9rem,4vw,3.4rem)] mt-3 text-balance">
            <Trans i18nKey="landing:features.heading" components={{ em: <em /> }} />
          </h2>
        </div>

        <div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-px rounded-2xl overflow-hidden border"
          style={{ borderColor: "hsl(var(--border))", background: "hsl(var(--border))" }}
        >
          {items.map(({ title, desc }, i) => {
            const Icon = ICONS[i] ?? Brain;
            return (
              <div
                key={title}
                className="p-8 sm:p-9 min-h-[260px] flex flex-col gap-4 transition-colors hover:bg-card/50"
                style={{ background: "hsl(var(--background))" }}
              >
                <div
                  className="h-10 w-10 rounded-lg grid place-items-center"
                  style={{
                    border: "1px solid hsl(var(--eop-line-strong))",
                    color: "hsl(var(--accent))",
                    background: "hsl(var(--eop-accent-soft))",
                  }}
                >
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="font-sans font-medium tracking-tight text-[19px] leading-tight">
                  {title}
                </h3>
                <p className="text-[14px] text-muted-foreground leading-relaxed">{desc}</p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
