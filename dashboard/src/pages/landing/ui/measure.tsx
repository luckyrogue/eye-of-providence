import { useTranslation } from "react-i18next";
import { Activity, BarChart3, Brain, Code2, GitBranch, Sparkles } from "lucide-react";
type Card = {
  title: string;
  body: string;
  pills: string[];
};
const ICONS = [Activity, Code2, Brain, GitBranch, BarChart3, Sparkles];
export function Measure() {
  const { t } = useTranslation("landing");
  const cards = t("measure.cards", { returnObjects: true }) as Card[];
  return (
    <section id="measure" className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px]">
        <div className="flex flex-col gap-6 mb-[60px] max-w-[800px]">
          <span className="eyebrow">{t("measure.eyebrow")}</span>
          <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] text-balance">
            {t("measure.title1")}
            <em>{t("measure.titleItal")}</em>
            {t("measure.title2")}
          </h2>
          <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
            {t("measure.sub")}
          </p>
        </div>
        <div
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-px rounded-2xl overflow-hidden border"
          style={{
            background: "hsl(var(--border))",
            borderColor: "hsl(var(--border))",
          }}
        >
          {cards.map((c, i) => {
            const Icon = ICONS[i] ?? Activity;
            return (
              <div
                key={c.title}
                className="px-[30px] py-9 min-h-[280px] flex flex-col gap-3.5 transition-colors hover:bg-card/40"
                style={{ background: "hsl(var(--background))" }}
              >
                <div
                  className="h-[38px] w-[38px] rounded-lg grid place-items-center"
                  style={{
                    border: "1px solid hsl(var(--eop-line-strong))",
                    color: "hsl(var(--accent))",
                    background: "hsl(var(--eop-accent-soft))",
                  }}
                >
                  <Icon className="h-[18px] w-[18px]" />
                </div>
                <h3 className="font-sans font-medium tracking-[-0.015em] text-[19px] leading-tight">
                  {c.title}
                </h3>
                <p className="text-muted-foreground text-[14px] leading-[1.6]">{c.body}</p>
                <div className="flex flex-wrap gap-1.5 mt-auto">
                  {c.pills.map((p, j) => (
                    <span key={p} className={`pill ${j === 0 ? "accent" : ""}`}>
                      {p}
                    </span>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
