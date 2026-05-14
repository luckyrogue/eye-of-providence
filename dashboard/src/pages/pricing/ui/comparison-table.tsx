import { Fragment } from "react";
import { Check, Minus } from "lucide-react";
import { Eyebrow } from "@eop/ui";
type Cell = string | boolean;
type Row = {
  label: string;
  values: [Cell, Cell, Cell, Cell];
};
type Group = {
  name: string;
  rows: Row[];
};
const TIERS = ["Free", "Pro", "Business", "Enterprise"] as const;
const GROUPS: Group[] = [
  {
    name: "Teams & users",
    rows: [
      { label: "Teams per account", values: ["1", "Unlimited", "Unlimited", "Unlimited"] },
      { label: "Users per team", values: ["5", "50", "Unlimited", "Unlimited"] },
      { label: "Roles (owner / admin / member)", values: [true, true, true, true] },
      { label: "Custom roles", values: [false, false, true, true] },
      { label: "Invite via link or email", values: [true, true, true, true] },
    ],
  },
  {
    name: "Data & retention",
    rows: [
      { label: "Event history", values: ["30 days", "365 days", "Unlimited", "Unlimited"] },
      { label: "Export to CSV / JSON", values: [true, true, true, true] },
      { label: "Per-project breakdown", values: [true, true, true, true] },
      { label: "Custom retention policy", values: [false, false, true, true] },
    ],
  },
  {
    name: "Insights & reporting",
    rows: [
      { label: "Live dashboard", values: [true, true, true, true] },
      { label: "AI-generated weekly reports", values: [false, true, true, true] },
      { label: "Anomaly detection", values: [false, true, true, true] },
      { label: "Email digests", values: [false, true, true, true] },
      { label: "Custom report schedules", values: [false, false, true, true] },
    ],
  },
  {
    name: "Integrations",
    rows: [
      { label: "Webhooks", values: ["1", "Unlimited", "Unlimited", "Unlimited"] },
      { label: "Webhook signing + retries", values: [false, false, true, true] },
      { label: "API tokens", values: ["1", "Standard limits", "Higher limits", "Custom"] },
      { label: "Browser extension", values: [true, true, true, true] },
      { label: "VSCode / Claude Code agent", values: [true, true, true, true] },
    ],
  },
  {
    name: "Security & compliance",
    rows: [
      { label: "Email + password auth", values: [true, true, true, true] },
      { label: "GitHub OAuth", values: [true, true, true, true] },
      { label: "SSO — OIDC", values: [false, false, true, true] },
      { label: "SSO — SAML 2.0", values: [false, false, true, true] },
      { label: "Audit log", values: [false, false, true, true] },
      { label: "SOC2 / GDPR reports", values: [false, false, false, true] },
      { label: "Custom security review", values: [false, false, false, true] },
    ],
  },
  {
    name: "Deployment & support",
    rows: [
      { label: "Cloud (managed)", values: [true, true, true, true] },
      { label: "Self-hosted Docker", values: [false, false, true, true] },
      { label: "On-prem deployment", values: [false, false, false, true] },
      { label: "Community support", values: [true, true, true, true] },
      { label: "Email support (48h)", values: [false, true, true, true] },
      { label: "Priority support (24h)", values: [false, false, true, true] },
      { label: "Dedicated success manager", values: [false, false, false, true] },
      { label: "99.9% SLA", values: [false, false, false, true] },
    ],
  },
];
function CellContent({ value }: { value: Cell }) {
  if (value === true) {
    return <Check className="h-4 w-4 text-emerald-600 dark:text-emerald-400" aria-label="Yes" />;
  }
  if (value === false) {
    return <Minus className="h-4 w-4 text-muted-foreground/40" aria-label="No" />;
  }
  return <span className="text-sm">{value}</span>;
}
export function ComparisonTable() {
  return (
    <section className="py-24 border-t bg-muted/20">
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <div className="text-center mb-12">
          <Eyebrow>Compare</Eyebrow>
          <h2 className="display-head text-3xl sm:text-4xl md:text-5xl mt-3">
            Every feature, side by side
          </h2>
          <p className="text-muted-foreground mt-4 max-w-xl mx-auto">
            All paid features are free during beta. Pricing kicks in at GA in Q4 2026.
          </p>
        </div>

        <div className="overflow-x-auto rounded-xl border bg-card">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-card/95 backdrop-blur z-10">
              <tr className="border-b">
                <th className="text-left font-medium text-muted-foreground px-4 py-4 w-[34%]">
                  Feature
                </th>
                {TIERS.map((tier, i) => (
                  <th
                    key={tier}
                    className={
                      "text-center font-display font-bold text-base px-4 py-4 " +
                      (i === 1 ? "text-foreground" : "text-muted-foreground")
                    }
                  >
                    {tier}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {GROUPS.map((group) => (
                <Fragment key={group.name}>
                  <tr className="bg-muted/30">
                    <td
                      colSpan={5}
                      className="px-4 py-2.5 text-[11px] font-mono uppercase tracking-wider text-muted-foreground"
                    >
                      {group.name}
                    </td>
                  </tr>
                  {group.rows.map((row) => (
                    <tr
                      key={group.name + "-" + row.label}
                      className="border-b last:border-b-0 hover:bg-muted/20"
                    >
                      <td className="px-4 py-3 text-sm">{row.label}</td>
                      {row.values.map((v, i) => (
                        <td key={i} className="px-4 py-3 text-center">
                          <div className="inline-flex items-center justify-center">
                            <CellContent value={v} />
                          </div>
                        </td>
                      ))}
                    </tr>
                  ))}
                </Fragment>
              ))}
            </tbody>
          </table>
        </div>

        <p className="mt-8 text-center text-sm text-muted-foreground">
          Have an unusual requirement?{" "}
          <a
            href="mailto:sales@eop.rysdavletov.org"
            className="text-foreground underline underline-offset-4"
          >
            Talk to us
          </a>
          .
        </p>
      </div>
    </section>
  );
}
