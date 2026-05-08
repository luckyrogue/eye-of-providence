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
      <div className="flex items-center justify-between gap-4">
        <div className="text-sm">
          <div className="font-medium">{title}</div>
          {description && <p className="text-xs text-muted-foreground mt-0.5">{description}</p>}
        </div>
        {action}
      </div>
    </div>
  );
}
