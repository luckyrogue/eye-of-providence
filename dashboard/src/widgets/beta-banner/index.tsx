import { Sparkles } from "lucide-react";
import type { BetaInfo } from "../../entities/team";

export function BetaBanner({ beta }: { beta: BetaInfo }) {
  const slotsLeft = beta.slots_remaining;
  const betaFull = beta.limit > 0 && slotsLeft === 0;
  return (
    <div className="rounded-xl border bg-gradient-to-br from-purple-500/5 to-blue-500/5 p-4 flex items-center justify-between gap-4">
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-lg bg-foreground/5 flex items-center justify-center">
          <Sparkles className="h-5 w-5 text-purple-500" />
        </div>
        <div>
          <div className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground">
            Beta program
          </div>
          <div className="text-sm font-medium mt-0.5">
            {betaFull
              ? `Все ${beta.limit} мест заняты — open seats coming soon`
              : `${slotsLeft} из ${beta.limit} мест свободно`}
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
