import { Trans, useTranslation } from "react-i18next";
import { Eyebrow, cn } from "@eop/ui";
import { Check } from "lucide-react";

type Tier = {
  name: string;
  price: string;
  period: string;
  features: string[];
  cta: string;
};

// "Pro" tier (index 1) — most-popular highlight (orange border + glow).
// Last tier ("Enterprise") routes to sales contact; everything else — signup.
function PriceCard({
  tier,
  index,
  total,
  highlight,
  mostPopular,
  betaBadge,
  isPaid,
}: {
  tier: Tier;
  index: number;
  total: number;
  highlight: boolean;
  mostPopular: string;
  betaBadge: string;
  isPaid: boolean;
}) {
  const { name, price, period, features, cta } = tier;
  const isEnterprise = index === total - 1;
  return (
    <div
      className={cn(
        "relative rounded-2xl p-7 flex flex-col transition-transform duration-300 hover:-translate-y-0.5",
      )}
      style={{
        border: highlight
          ? "1px solid hsl(var(--accent))"
          : "1px solid hsl(var(--eop-line-strong))",
        background: "hsl(var(--card) / 0.6)",
        boxShadow: highlight ? "0 0 60px -20px hsl(var(--eop-accent-glow))" : undefined,
      }}
    >
      {highlight && (
        <span
          className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full px-3 py-1 text-[10px] font-mono uppercase tracking-widest2"
          style={{
            background: "hsl(var(--accent))",
            color: "hsl(var(--accent-foreground))",
            boxShadow: "0 0 20px hsl(var(--eop-accent-glow))",
          }}
        >
          {mostPopular}
        </span>
      )}
      <h3 className="font-sans font-medium text-[22px] tracking-tight">{name}</h3>
      <div className="mt-4 flex items-baseline gap-1">
        <span className="font-sans text-[44px] font-medium tracking-tightest tabular-nums leading-none">
          {price}
        </span>
        <span className="font-mono text-[12px] text-muted-foreground ml-2">{period}</span>
      </div>
      {isPaid && (
        <span
          className="mt-3 inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-[10px] font-mono uppercase tracking-widest2"
          style={{
            background: "hsl(var(--eop-accent-soft))",
            color: "hsl(var(--accent))",
            border: "1px solid hsl(var(--accent) / 0.4)",
          }}
        >
          <span
            className="h-1.5 w-1.5 rounded-sm"
            style={{ background: "hsl(var(--accent))" }}
            aria-hidden
          />
          {betaBadge}
        </span>
      )}
      <ul className="mt-6 space-y-3 text-[13.5px] flex-1">
        {features.map((f) => (
          <li key={f} className="flex items-start gap-2.5 text-muted-foreground">
            <Check className="h-4 w-4 mt-0.5 text-[hsl(var(--accent))] shrink-0" />
            <span className="text-foreground/85">{f}</span>
          </li>
        ))}
      </ul>
      <a
        href={isEnterprise ? "mailto:sales@eop.rysdavletov.org" : "/dashboard"}
        className={cn(
          "mt-7 w-full inline-flex items-center justify-center rounded-lg px-4 py-2.5 text-[14px] font-medium transition-colors",
          highlight ? "btn-eop-primary" : "border hover:bg-foreground/5",
        )}
        style={
          highlight
            ? undefined
            : {
                borderColor: "hsl(var(--eop-line-strong))",
                color: "hsl(var(--foreground))",
              }
        }
      >
        {cta}
      </a>
    </div>
  );
}

export function Pricing() {
  const { t } = useTranslation("landing");
  const tiers = t("pricing.tiers", { returnObjects: true }) as Tier[];
  return (
    <section id="pricing" className="py-24 sm:py-32 px-4 sm:px-8">
      <div className="mx-auto max-w-7xl">
        <div className="text-center mb-14 sm:mb-16">
          <Eyebrow>{t("pricing.eyebrow")}</Eyebrow>
          <h2 className="display-head text-[clamp(1.9rem,4vw,3.4rem)] mt-3 max-w-2xl mx-auto text-balance">
            <Trans i18nKey="landing:pricing.heading" components={{ em: <em /> }} />
          </h2>
          <p className="text-muted-foreground mt-5 max-w-[58ch] mx-auto text-[15px] text-pretty">
            {t("pricing.lead")}
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 max-w-7xl mx-auto">
          {tiers.map((tier, i) => (
            <PriceCard
              key={tier.name}
              tier={tier}
              index={i}
              total={tiers.length}
              highlight={i === 1}
              mostPopular={t("pricing.most_popular")}
              betaBadge={t("pricing.beta_badge")}
              isPaid={i > 0 && i < tiers.length - 1}
            />
          ))}
        </div>
        <div className="mt-10 text-center">
          <a
            href="/pricing"
            className="text-sm font-medium text-foreground/80 hover:text-foreground underline underline-offset-4"
          >
            {t("pricing.compare_link")}
          </a>
        </div>
      </div>
    </section>
  );
}
