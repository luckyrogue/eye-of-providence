import { cn } from "./cn";

export type Plan = "free" | "pro" | "team" | "enterprise" | string;

const planColors: Record<string, string> = {
  free: "bg-muted text-muted-foreground border-border",
  pro: "bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30",
  team: "bg-purple-500/10 text-purple-700 dark:text-purple-300 border-purple-500/30",
  enterprise: "bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30",
};

export function PlanBadge({
  plan,
  until,
  className,
  untilLabel = "until",
  expiredLabel = "expired",
}: {
  plan: Plan;
  until?: string | null;
  className?: string;
  // i18n: caller passes locale-aware strings; English defaults для consumer'ов
  // которые не настроили i18n.
  untilLabel?: string;
  expiredLabel?: string;
}) {
  const dueDate = until ? new Date(until) : null;
  const isActive = plan !== "free" && (!dueDate || dueDate > new Date());
  const colorClass = planColors[plan] ?? planColors.free;
  return (
    <div className={cn("inline-flex items-center gap-2", className)}>
      <span
        className={cn(
          "inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-mono uppercase tracking-widest2",
          colorClass,
        )}
      >
        {plan}
      </span>
      {dueDate && (
        <span
          className={cn(
            "text-[10px] font-mono",
            isActive ? "text-muted-foreground" : "text-destructive",
          )}
        >
          {isActive ? untilLabel : expiredLabel} {dueDate.toISOString().slice(0, 10)}
        </span>
      )}
    </div>
  );
}
