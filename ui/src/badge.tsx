import type { ReactNode } from "react";
import { cn } from "./cn";

export type BadgeTone = "neutral" | "blue" | "purple" | "amber" | "green" | "red";

const tones: Record<BadgeTone, string> = {
  neutral: "bg-muted text-muted-foreground border-border",
  blue: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30",
  purple: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/30",
  amber: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30",
  green: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30",
  red: "bg-destructive/10 text-destructive border-destructive/30",
};

export function Badge({
  tone = "neutral",
  mono,
  children,
  className,
}: {
  tone?: BadgeTone;
  mono?: boolean;
  children: ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium",
        mono && "font-mono uppercase tracking-widest2",
        tones[tone],
        className,
      )}
    >
      {children}
    </span>
  );
}
