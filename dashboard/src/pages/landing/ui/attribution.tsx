import { useTranslation } from "react-i18next";

type StreamRow = { ts: string; tag: string; msg: string };

export function Attribution() {
  const { t } = useTranslation("landing");
  const stream = t("attribution.stream", { returnObjects: true }) as StreamRow[];
  return (
    <section id="attribution" className="py-[120px] px-4 sm:px-8 relative z-[2]">
      <div className="mx-auto max-w-[1200px] grid grid-cols-1 lg:grid-cols-[1fr_1.4fr] gap-10 lg:gap-[60px] items-center">
        <div className="flex flex-col gap-5">
          <span className="eyebrow">{t("attribution.eyebrow")}</span>
          <h2 className="font-display font-medium text-[clamp(1.875rem,4.4vw,3.5rem)] leading-[1.05] tracking-[-0.03em] text-balance">
            {t("attribution.title1")}
            <em>{t("attribution.titleItal")}</em>
            {t("attribution.title2")}
          </h2>
          <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.9375rem,1.4vw,1.125rem)]">
            {t("attribution.sub")}
          </p>
          <div className="flex flex-wrap gap-2 mt-2">
            <span className="tag typed">typed</span>
            <span className="tag ai-inline">ai-inline</span>
            <span className="tag ai-agent">ai-agent</span>
            <span className="tag paste-ai">pasted-ai</span>
            <span className="tag refactor">refactor</span>
          </div>
        </div>

        <div
          className="rounded-xl overflow-hidden font-mono text-[12px]"
          style={{
            border: "1px solid hsl(var(--eop-line-strong))",
            background: "hsl(var(--card))",
          }}
        >
          <div
            className="flex justify-between px-3.5 py-2.5"
            style={{
              borderBottom: "1px solid hsl(var(--border))",
              color: "hsl(var(--muted-foreground))",
              fontSize: 11,
            }}
          >
            <span>~/eop/attribution.log</span>
            <span>tail -f</span>
          </div>
          <div className="flex flex-col">
            {stream.map((r, i) => (
              <div
                key={i}
                className="grid grid-cols-[60px_1fr_110px] px-3.5 py-2.5 items-center gap-2.5"
                style={{
                  borderBottom: i === stream.length - 1 ? "none" : "1px solid hsl(var(--border))",
                  animation: `eop-slide-in 0.4s ease both`,
                  animationDelay: `${i * 0.08}s`,
                }}
              >
                <span className="text-muted-foreground text-[11px]">{r.ts}</span>
                <span className="text-[12px]">{r.msg}</span>
                <span className={`tag ${r.tag}`}>{r.tag}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
