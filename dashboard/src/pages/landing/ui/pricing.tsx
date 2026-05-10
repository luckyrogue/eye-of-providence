import { Trans, useTranslation } from "react-i18next";
import { Button, Eyebrow, cn } from "@eop/ui";
import { Check } from "lucide-react";

type Tier = {
  name: string;
  price: string;
  period: string;
  features: string[];
  cta: string;
};

// highlight middle tier (Founding Company) — index === 1
function PriceCard({ tier, highlight, mostPopular }: { tier: Tier; highlight: boolean; mostPopular: string }) {
  const { name, price, period, features, cta } = tier;
  return (
    <div className={cn("relative rounded-xl border p-7 card-hover", highlight ? "border-foreground bg-card shadow-lg" : "bg-card")}>
      {highlight && (
        <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-foreground text-background px-3 py-1 text-[10px] font-mono uppercase tracking-widest2">
          {mostPopular}
        </span>
      )}
      <h3 className="font-display font-bold text-2xl tracking-tight">{name}</h3>
      <div className="mt-4 flex items-baseline gap-1">
        <span className="font-display text-4xl sm:text-5xl font-bold tracking-tightest tabular-nums">{price}</span>
        <span className="text-sm text-muted-foreground ml-2">{period}</span>
      </div>
      <ul className="mt-6 space-y-2.5 text-sm">
        {features.map((f) => (
          <li key={f} className="flex items-start gap-2">
            <Check className="h-4 w-4 mt-0.5 text-foreground shrink-0" />
            <span>{f}</span>
          </li>
        ))}
      </ul>
      <Button asChild className="w-full mt-7" variant={highlight ? "default" : "outline"}>
        <a href="/dashboard">{cta}</a>
      </Button>
    </div>
  );
}

export function Pricing() {
  const { t } = useTranslation("landing");
  const tiers = t("pricing.tiers", { returnObjects: true }) as Tier[];
  return (
    <section id="pricing" className="py-24">
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <div className="text-center mb-12">
          <Eyebrow>{t("pricing.eyebrow")}</Eyebrow>
          <h2 className="display-head text-3xl sm:text-4xl md:text-5xl mt-3 max-w-2xl mx-auto">
            <Trans i18nKey="landing:pricing.heading" components={{ em: <em /> }} />
          </h2>
          <p className="text-muted-foreground mt-4 max-w-lg mx-auto">{t("pricing.lead")}</p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-5xl mx-auto">
          {tiers.map((tier, i) => (
            <PriceCard
              key={tier.name}
              tier={tier}
              highlight={i === 1}
              mostPopular={t("pricing.most_popular")}
            />
          ))}
        </div>
      </div>
    </section>
  );
}
