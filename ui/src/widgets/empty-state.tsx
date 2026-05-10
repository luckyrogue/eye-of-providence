import type { ReactNode } from "react";
import { cn } from "../shared/lib/cn";

export function EmptyState({
  eyebrow,
  title,
  description,
  action,
  className,
}: {
  eyebrow?: string;
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("py-12 text-center space-y-3", className)}>
      {eyebrow && (
        <div className="font-mono text-[10px] uppercase tracking-widest3 text-muted-foreground">
          {eyebrow}
        </div>
      )}
      <div className="font-display tracking-tight text-base">{title}</div>
      {description && <p className="text-sm text-muted-foreground max-w-md mx-auto">{description}</p>}
      {action && <div className="pt-2">{action}</div>}
    </div>
  );
}
