import type { ReactNode } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "../shared/ui/card";
import { cn } from "../shared/lib/cn";

export type StatTileAccent = "purple" | "blue" | "amber" | "neutral";

const accentClasses: Record<StatTileAccent, string> = {
  purple: "from-purple-500/20",
  blue: "from-blue-500/20",
  amber: "from-amber-500/20",
  neutral: "from-foreground/10",
};

export function StatTile({
  label,
  value,
  unit,
  hint,
  icon,
  accent = "neutral",
  className,
}: {
  label: string;
  value: string | number;
  unit?: string;
  hint?: ReactNode;
  icon?: ReactNode;
  accent?: StatTileAccent;
  className?: string;
}) {
  return (
    <Card className={cn("overflow-hidden relative card-hover", className)}>
      <div
        className={cn(
          "absolute right-0 top-0 h-24 w-24 rounded-bl-full bg-gradient-to-bl to-transparent",
          accentClasses[accent],
        )}
      />
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">
            {label}
          </CardTitle>
          {icon}
        </div>
      </CardHeader>
      <CardContent>
        <div className="font-display text-5xl font-bold tabular-nums tracking-tightest">
          {value}
          {unit && <span className="text-base font-normal text-muted-foreground"> {unit}</span>}
        </div>
        {hint && <p className="text-xs text-muted-foreground mt-2 font-mono">{hint}</p>}
      </CardContent>
    </Card>
  );
}
