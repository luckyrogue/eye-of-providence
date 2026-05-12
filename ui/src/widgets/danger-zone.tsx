import type { ReactNode } from "react";

export function DangerZone({
  title,
  description,
  action,
}: {
  title: ReactNode;
  description?: ReactNode;
  action: ReactNode;
}) {
  return (
    <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-4">
      <div className="font-mono text-[11px] uppercase tracking-widest2 text-destructive mb-1">
        Danger zone
      </div>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
        <div className="min-w-0 flex-1 text-sm">
          <div className="font-medium">{title}</div>
          {description && <p className="text-xs text-muted-foreground mt-0.5">{description}</p>}
        </div>
        <div className="w-full shrink-0 sm:w-auto [&_button]:w-full sm:[&_button]:w-auto">
          {action}
        </div>
      </div>
    </div>
  );
}
