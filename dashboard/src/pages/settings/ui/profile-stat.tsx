import type { ReactNode } from "react";
export function ProfileStat({
  icon,
  label,
  value,
}: {
  icon: ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="rounded-md border p-3 flex items-start gap-3">
      <div className="text-muted-foreground mt-0.5">{icon}</div>
      <div className="min-w-0 flex-1">
        <div className="text-[11px] uppercase tracking-widest3 text-muted-foreground font-mono">
          {label}
        </div>
        <div className="text-sm font-medium truncate mt-0.5">{value}</div>
      </div>
    </div>
  );
}
