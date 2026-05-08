import { Button, Eyebrow, cn } from "@eop/ui";
import { Check } from "lucide-react";

type Tier = {
  name: string;
  price: string;
  period: string;
  features: string[];
  highlight: boolean;
  cta: string;
};

const TIERS: Tier[] = [
  {
    name: "Solo",
    price: "Free",
    period: "forever",
    highlight: false,
    cta: "Get started",
    features: ["Personal dashboard", "Up to 3 projects", "30-day event history", "Weekly AI report", "Community support"],
  },
  {
    name: "Founding Company",
    price: "Free",
    period: "for the first 3",
    highlight: true,
    cta: "Claim a slot",
    features: [
      "Everything in Solo",
      "Unlimited team members",
      "Unlimited projects",
      "Roles, invites, member analytics",
      "18-month event history",
      "Direct line to the founders",
    ],
  },
  {
    name: "Self-hosted",
    price: "Free",
    period: "open-core",
    highlight: false,
    cta: "Read the docs",
    features: ["Run the whole stack", "Your data on your infra", "Docker / docker-compose", "Postgres + ClickHouse", "Community support"],
  },
];

function PriceCard({ tier }: { tier: Tier }) {
  const { name, price, period, features, highlight, cta } = tier;
  return (
    <div className={cn("relative rounded-xl border p-7 card-hover", highlight ? "border-foreground bg-card shadow-lg" : "bg-card")}>
      {highlight && (
        <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-foreground text-background px-3 py-1 text-[10px] font-mono uppercase tracking-widest2">
          Most popular
        </span>
      )}
      <h3 className="font-display font-bold text-2xl tracking-tight">{name}</h3>
      <div className="mt-4 flex items-baseline gap-1">
        <span className="font-display text-5xl font-bold tracking-tightest tabular-nums">{price}</span>
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
  return (
    <section id="pricing" className="py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="text-center mb-12">
          <Eyebrow>Pricing · Beta</Eyebrow>
          <h2 className="display-head text-4xl md:text-5xl mt-3 max-w-2xl mx-auto">
            Free for the <em>first 3 companies</em>. Forever for solo devs.
          </h2>
          <p className="text-muted-foreground mt-4 max-w-lg mx-auto">
            We onboard 3 founding companies free of charge — no credit card, no trial countdown, no seat caps. Self-hosted is free for everyone.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-5xl mx-auto">
          {TIERS.map((t) => (
            <PriceCard key={t.name} tier={t} />
          ))}
        </div>
      </div>
    </section>
  );
}
