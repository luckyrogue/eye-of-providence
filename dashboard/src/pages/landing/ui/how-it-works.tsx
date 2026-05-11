// How it works — 4 step cards с inline SVG-диаграммой на каждом (artifact 1:1).

import { useTranslation } from "react-i18next";

type Step = { num: string; title: string; body: string };

function HowDiagram({ kind }: { kind: "observe" | "categorize" | "attribute" | "synthesize" }) {
  const stroke = "rgba(255,255,255,0.5)";
  const acc = "hsl(var(--accent))";
  switch (kind) {
    case "observe":
      return (
        <svg viewBox="0 0 200 80" width="100%" height="80">
          <rect x="10" y="20" width="40" height="40" rx="4" fill="none" stroke={stroke} />
          <text x="30" y="44" fill={stroke} fontSize="9" fontFamily="monospace" textAnchor="middle">
            OS
          </text>
          <path d="M50 40 L 90 40" stroke={stroke} strokeWidth="1" strokeDasharray="2 2" />
          <circle cx="120" cy="40" r="14" fill="none" stroke={acc} strokeWidth="1.5" />
          <circle cx="120" cy="40" r="3" fill={acc} />
          <path d="M150 40 L 190 40" stroke={stroke} strokeWidth="1" strokeDasharray="2 2" />
          <rect x="180" y="25" width="14" height="30" fill="none" stroke={stroke} />
        </svg>
      );
    case "categorize":
      return (
        <svg viewBox="0 0 200 80" width="100%" height="80">
          {[0, 1, 2, 3, 4].map((i) => (
            <rect
              key={i}
              x={20 + i * 32}
              y={i % 2 ? 20 : 40}
              width="24"
              height="20"
              rx="2"
              fill="none"
              stroke={i === 2 ? acc : stroke}
            />
          ))}
          <path
            d="M20 70 Q 100 70 180 30"
            fill="none"
            stroke={acc}
            strokeWidth="1.5"
            strokeDasharray="3 3"
          />
        </svg>
      );
    case "attribute":
      return (
        <svg viewBox="0 0 200 80" width="100%" height="80">
          {Array.from({ length: 14 }).map((_, i) => {
            const palette = [acc, "#4ade80", "#60a5fa", "#c084fc"];
            return (
              <rect
                key={i}
                x={10 + i * 13}
                y={20 + (i % 3) * 8}
                width="10"
                height="6"
                fill={palette[i % 4]}
                opacity={0.5 + (i % 3) * 0.2}
              />
            );
          })}
          <path d="M10 60 H 190" stroke={stroke} />
        </svg>
      );
    case "synthesize":
      return (
        <svg viewBox="0 0 200 80" width="100%" height="80">
          <path
            d="M20 60 Q 50 50 70 45 Q 100 38 130 25 Q 160 18 180 12"
            fill="none"
            stroke={acc}
            strokeWidth="1.5"
          />
          <path
            d="M20 60 Q 50 55 70 52 Q 100 48 130 42 Q 160 38 180 35"
            fill="none"
            stroke={stroke}
            strokeWidth="1"
            strokeDasharray="2 2"
          />
          <circle cx="180" cy="12" r="3" fill={acc} />
        </svg>
      );
  }
}

const KINDS = ["observe", "categorize", "attribute", "synthesize"] as const;

export function HowItWorks() {
  const { t } = useTranslation("landing");
  const steps = t("how.steps", { returnObjects: true }) as Step[];
  return (
    <section id="how" className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px]">
        <div className="flex flex-col gap-5 max-w-[720px]">
          <span className="eyebrow">{t("how.eyebrow")}</span>
          <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] text-balance">
            {t("how.title")}
          </h2>
          <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
            {t("how.sub")}
          </p>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-[60px]">
          {steps.map((s, i) => (
            <div
              key={s.title}
              className="p-6 rounded-xl border flex flex-col gap-3.5 min-h-[220px]"
              style={{
                borderColor: "hsl(var(--border))",
                background: "rgba(255,255,255,0.01)",
              }}
            >
              <span
                className="font-mono text-[11px] tracking-[0.15em]"
                style={{ color: "hsl(var(--accent))" }}
              >
                {s.num}
              </span>
              <h4 className="font-sans font-medium text-[18px] tracking-[-0.015em]">{s.title}</h4>
              <p className="text-[13px] text-muted-foreground leading-[1.6]">{s.body}</p>
              <div className="h-20 mt-auto opacity-90">
                <HowDiagram kind={KINDS[i] ?? "observe"} />
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
