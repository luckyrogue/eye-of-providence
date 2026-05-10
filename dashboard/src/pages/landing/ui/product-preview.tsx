import { useTranslation } from "react-i18next";
import { cn } from "@eop/ui";

const ACCENTS = {
  purple: "from-purple-500/20",
  blue: "from-blue-500/20",
  amber: "from-amber-500/20",
} as const;

function PreviewStat({ label, value, accent }: { label: string; value: string; accent: keyof typeof ACCENTS }) {
  return (
    <div className="rounded-lg border bg-card p-2.5 sm:p-3 relative overflow-hidden">
      <div className={cn("absolute right-0 top-0 h-12 w-12 sm:h-16 sm:w-16 rounded-bl-full bg-gradient-to-bl to-transparent", ACCENTS[accent])} />
      <div className="font-mono text-[9px] sm:text-[10px] uppercase tracking-widest2 text-muted-foreground line-clamp-2">{label}</div>
      <div className="font-display text-xl sm:text-2xl md:text-3xl font-bold tracking-tightest tabular-nums mt-1 truncate">{value}</div>
    </div>
  );
}

function FakeChart() {
  // Cosmetic SVG для preview-карточки на лендинге.
  const points = [10, 22, 18, 30, 26, 38, 34, 48, 42, 56, 50, 62, 58, 70, 66, 76];
  const pointsAi = [4, 8, 6, 10, 12, 14, 16, 22, 26, 30, 34, 36, 40, 44, 48, 52];
  const max = 80;
  const path = (arr: number[]) =>
    arr.map((v, i) => `${(i / (arr.length - 1)) * 100},${100 - (v / max) * 90}`).join(" ");
  return (
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="absolute inset-0 w-full h-full mt-6 px-3 pb-3">
      <defs>
        <linearGradient id="manualG" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="hsl(220 80% 55%)" stopOpacity="0.3" />
          <stop offset="100%" stopColor="hsl(220 80% 55%)" stopOpacity="0" />
        </linearGradient>
        <linearGradient id="aiG" x1="0" x2="0" y1="0" y2="1">
          <stop offset="0%" stopColor="hsl(280 70% 60%)" stopOpacity="0.3" />
          <stop offset="100%" stopColor="hsl(280 70% 60%)" stopOpacity="0" />
        </linearGradient>
      </defs>
      <polyline points={path(points)} fill="none" stroke="hsl(220 80% 55%)" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
      <polyline points={path(pointsAi)} fill="none" stroke="hsl(280 70% 60%)" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
      <polygon points={`0,100 ${path(points)} 100,100`} fill="url(#manualG)" />
      <polygon points={`0,100 ${path(pointsAi)} 100,100`} fill="url(#aiG)" />
    </svg>
  );
}

export function ProductPreview() {
  const { t } = useTranslation("landing");
  return (
    <div className="relative mt-16 mx-auto max-w-5xl reveal reveal-delay-5">
      <div className="absolute -inset-x-8 -inset-y-4 bg-gradient-to-b from-transparent via-purple-500/5 to-blue-500/10 blur-2xl pointer-events-none" />
      <div className="relative rounded-xl border bg-card/80 backdrop-blur-sm shadow-2xl overflow-hidden">
        <div className="border-b px-4 py-2 flex items-center gap-2 bg-muted/30">
          <span className="h-2.5 w-2.5 rounded-full bg-red-500/70" />
          <span className="h-2.5 w-2.5 rounded-full bg-amber-500/70" />
          <span className="h-2.5 w-2.5 rounded-full bg-emerald-500/70" />
          <span className="ml-3 font-mono text-xs text-muted-foreground">eop.rysdavletov.org/dashboard</span>
        </div>
        <div className="p-6 space-y-4">
          <div className="grid grid-cols-3 gap-3">
            <PreviewStat label={t("preview.ai_ratio")} value="42%" accent="purple" />
            <PreviewStat label={t("preview.active_time")} value="6h 18m" accent="blue" />
            <PreviewStat label={t("preview.manual")} value="58%" accent="amber" />
          </div>
          <div className="rounded-lg border bg-card p-4 h-44 relative overflow-hidden">
            <span className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">
              {t("preview.last_30")}
            </span>
            <FakeChart />
          </div>
        </div>
      </div>
    </div>
  );
}
