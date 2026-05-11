// Hero: split-grid (copy слева, 3D eye справа), hero-stats внизу 4-col grid
// с анимированным live-counter. Дизайн 1:1 с Claude artifact (Eye of Providence).

import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { ArrowRight, Download, Github, Search } from "lucide-react";
import { EyeScene } from "./eye-scene";

// StatNum — анимированный count-up с intersection-observer (трек на первое
// появление в viewport'е). Иначе число "прыгает" каждый раз когда секция
// скроллится в фокус.
function StatNum({ end, decimals = 0 }: { end: number; decimals?: number }) {
  const [val, setVal] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);
  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            const dur = 1400;
            const t0 = performance.now();
            const tick = (t: number) => {
              const p = Math.min(1, (t - t0) / dur);
              const eased = 1 - Math.pow(1 - p, 3);
              setVal(end * eased);
              if (p < 1) requestAnimationFrame(tick);
            };
            requestAnimationFrame(tick);
            io.disconnect();
          }
        }
      },
      { threshold: 0.3 },
    );
    io.observe(node);
    return () => io.disconnect();
  }, [end]);
  return (
    <span ref={ref}>{decimals ? val.toFixed(decimals) : Math.round(val).toLocaleString()}</span>
  );
}

type Stat = { label: string; value: string; unit: string };

export function Hero() {
  const { t } = useTranslation("landing");
  const stats = t("stats", { returnObjects: true }) as Stat[];

  return (
    <section className="relative grid grid-cols-1 lg:grid-cols-[1.1fr_1fr] items-center gap-10 lg:gap-[60px] pt-[120px] pb-[80px] px-4 sm:px-8 mx-auto max-w-[1400px]">
      <div className="flex flex-col gap-7">
        <span className="eyebrow">{t("hero.eyebrow")}</span>
        <h1 className="display font-display font-medium leading-[0.98] tracking-[-0.035em] text-balance text-[clamp(2.5rem,6.5vw,5.4rem)]">
          {t("hero.title1")}
          <span style={{ color: "hsl(var(--accent))" }}>{t("hero.titleAccent")}</span>
          {t("hero.title2")}
          <em>{t("hero.titleItal")}</em>
          {t("hero.title3")}
        </h1>
        <p className="text-muted-foreground max-w-[56ch] text-pretty text-[clamp(0.95rem,1.4vw,1.125rem)]">
          {t("hero.sub")}
        </p>
        <div className="flex flex-wrap gap-3 items-center">
          <a
            href="/dashboard"
            className="btn-eop-primary inline-flex items-center gap-2 px-[18px] py-[11px] rounded-lg text-[14px] font-medium"
          >
            <Download className="h-[18px] w-[18px]" />
            {t("hero.ctaPrimary")}
          </a>
          <a
            href="#how"
            className="inline-flex items-center gap-2 px-[18px] py-[11px] rounded-lg text-[14px] font-medium border transition-colors hover:bg-foreground/5"
            style={{
              borderColor: "hsl(var(--eop-line-strong))",
              background: "rgba(255,255,255,0.02)",
            }}
          >
            {t("hero.ctaSecondary")}
            <ArrowRight className="h-[18px] w-[18px]" />
          </a>
          <div
            className="flex gap-2 flex-wrap font-mono text-[11px]"
            style={{ color: "hsl(var(--muted-foreground))" }}
          >
            <span
              className="inline-flex items-center gap-1.5 border rounded-md px-2 py-1"
              style={{
                borderColor: "hsl(var(--border))",
                background: "rgba(255,255,255,0.02)",
              }}
            >
              <Search className="h-[11px] w-[11px]" />
              {t("hero.kbd")}
            </span>
          </div>
        </div>
        <div className="flex flex-wrap gap-5 font-mono text-[12px] tracking-[0.06em] text-muted-foreground">
          <span className="inline-flex items-center gap-1.5">
            <span
              className="h-1.5 w-1.5 rounded-full animate-pulse"
              style={{ background: "hsl(var(--success))" }}
            />
            {t("hero.meta1")}: 2,847
          </span>
          <span className="inline-flex items-center gap-1.5">{t("hero.meta2")}: 03/2026</span>
          <span className="inline-flex items-center gap-1.5">
            <Github className="h-3 w-3" />
            {t("hero.meta3")}
          </span>
        </div>
      </div>

      <EyeScene />

      {/* Hero stats: 4-col flat-grid, 1px lines между секциями */}
      <div
        className="col-span-full grid grid-cols-2 md:grid-cols-4 gap-px rounded-xl overflow-hidden border mt-10"
        style={{
          background: "hsl(var(--border))",
          borderColor: "hsl(var(--border))",
        }}
      >
        {stats.map((s) => (
          <div
            key={s.label}
            className="flex flex-col gap-1.5 px-5 py-4"
            style={{ background: "hsl(var(--background))" }}
          >
            <span className="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
              {s.label}
            </span>
            <span className="text-[22px] tabular-nums tracking-[-0.02em]">
              <StatNum end={parseFloat(s.value)} decimals={s.value.includes(".") ? 1 : 0} />
              <span className="text-muted-foreground text-[14px] ml-1.5">{s.unit}</span>
            </span>
          </div>
        ))}
      </div>
    </section>
  );
}
