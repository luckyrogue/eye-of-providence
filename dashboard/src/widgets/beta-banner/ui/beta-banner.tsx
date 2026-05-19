import { useTranslation } from "react-i18next";
import { Sparkles } from "lucide-react";
import type { BetaInfo } from "@/entities/team";

export function BetaBanner({ beta }: { beta: BetaInfo }) {
  const { t } = useTranslation("app");
  const slotsLeft = beta.slots_remaining;
  const betaFull = beta.limit > 0 && slotsLeft === 0;
  return (
    <div
      className="rounded-xl border p-4 flex items-center justify-between gap-4"
      style={{
        background: "linear-gradient(135deg, hsl(var(--accent) / 0.06), hsl(var(--accent) / 0.02))",
      }}
    >
      <div className="flex items-center gap-3">
        <div
          className="h-10 w-10 rounded-lg flex items-center justify-center"
          style={{
            background: "hsl(var(--accent) / 0.1)",
            border: "1px solid hsl(var(--accent) / 0.3)",
          }}
        >
          <Sparkles className="h-5 w-5" style={{ color: "hsl(var(--accent))" }} />
        </div>
        <div>
          <div className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground">
            {t("beta_banner.title")}
          </div>
          <div className="text-sm font-medium mt-0.5">
            {betaFull
              ? t("beta_banner.full", { limit: beta.limit })
              : t("beta_banner.free_count", { slots: slotsLeft, limit: beta.limit })}
          </div>
        </div>
      </div>
      <div className="font-display text-3xl font-bold tabular-nums tracking-tightest text-muted-foreground">
        {beta.teams_count}
        <span className="text-muted-foreground/50">/{beta.limit}</span>
      </div>
    </div>
  );
}
