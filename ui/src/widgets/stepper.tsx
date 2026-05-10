import { Check } from "lucide-react";
import { cn } from "../shared/lib/cn";

export type StepperItem = { key: string; label: string };

export function Stepper({
  steps,
  current,
  className,
}: {
  steps: StepperItem[];
  current: string;
  className?: string;
}) {
  const index = steps.findIndex((s) => s.key === current);
  return (
    <div className={cn("flex items-center gap-2", className)}>
      {steps.map((s, i) => {
        const isActive = i === index;
        const isDone = i < index;
        return (
          <div key={s.key} className="flex-1 flex items-center gap-2">
            <div
              className={cn(
                "flex items-center justify-center h-7 w-7 rounded-full text-[11px] font-mono transition-colors shrink-0",
                isDone
                  ? "bg-primary text-primary-foreground"
                  : isActive
                    ? "border-2 border-primary text-primary"
                    : "border border-border text-muted-foreground",
              )}
            >
              {isDone ? <Check className="h-3.5 w-3.5" /> : i + 1}
            </div>
            <span
              className={cn(
                "text-xs font-mono uppercase tracking-widest2 truncate",
                isActive || isDone ? "text-foreground" : "text-muted-foreground",
              )}
            >
              {s.label}
            </span>
            {i < steps.length - 1 && <div className="flex-1 h-px bg-border" />}
          </div>
        );
      })}
    </div>
  );
}
