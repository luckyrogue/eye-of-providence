import { useTranslation } from "react-i18next";
import { Shield, Lock, GitBranch, Users, Trash2 } from "lucide-react";

const ICONS = [Shield, Lock, GitBranch, Users, Trash2];

type Principle = { strong: string; body: string };

export function Privacy() {
  const { t } = useTranslation("landing");
  const principles = t("privacy.principles", { returnObjects: true }) as Principle[];
  const shieldSends = t("privacy.shieldSends", { returnObjects: true }) as string[];
  const shieldNoSends = t("privacy.shieldNoSends", { returnObjects: true }) as string[];

  return (
    <section id="privacy" className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px] grid grid-cols-1 lg:grid-cols-[1.2fr_1fr] gap-10 lg:gap-[60px] items-start">
        <div>
          <div className="flex flex-col gap-5 mb-8">
            <span className="eyebrow">{t("privacy.eyebrow")}</span>
            <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] text-balance">
              {t("privacy.title1")}
              <em>{t("privacy.titleItal")}</em>
            </h2>
            <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
              {t("privacy.sub")}
            </p>
          </div>
          <ul className="flex flex-col gap-4">
            {principles.map((p, i) => {
              const Icon = ICONS[i] ?? Shield;
              return (
                <li
                  key={p.strong}
                  className="grid grid-cols-[28px_1fr] gap-3.5 items-start py-3.5"
                  style={{
                    borderBottom:
                      i === principles.length - 1 ? "none" : "1px solid hsl(var(--border))",
                  }}
                >
                  <span
                    className="h-7 w-7 rounded-md grid place-items-center shrink-0"
                    style={{
                      border: "1px solid hsl(var(--eop-line-strong))",
                      color: "hsl(var(--accent))",
                    }}
                  >
                    <Icon className="h-3.5 w-3.5" />
                  </span>
                  <div>
                    <strong className="font-medium text-[14px] block mb-1">{p.strong}</strong>
                    <p className="text-[13px] text-muted-foreground">{p.body}</p>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        <div
          className="rounded-2xl p-7 font-mono text-[12px] leading-[1.8]"
          style={{
            border: "1px solid hsl(var(--eop-line-strong))",
            background: "linear-gradient(180deg, hsl(var(--accent) / 0.04), transparent)",
          }}
        >
          <h5
            className="font-sans text-[14px] mb-4 tracking-[0.05em] uppercase"
            style={{ color: "hsl(var(--muted-foreground))" }}
          >
            {t("privacy.shieldTitle")}
          </h5>
          {shieldSends.map((k) => (
            <div
              key={k}
              className="flex justify-between py-1"
              style={{ borderBottom: "1px dashed hsl(var(--border))" }}
            >
              <span>{k}</span>
              <span style={{ color: "hsl(var(--success))" }}>→ sent</span>
            </div>
          ))}
          <div className="h-4" />
          {shieldNoSends.map((k, i) => (
            <div
              key={k}
              className="flex justify-between py-1"
              style={{
                borderBottom:
                  i === shieldNoSends.length - 1 ? "none" : "1px dashed hsl(var(--border))",
              }}
            >
              <span>{k}</span>
              <span style={{ color: "#ef4444" }}>× never</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
