import type { ReactNode } from "react";
import { cn } from "../shared/lib/cn";

export function PageHeader({
  eyebrow,
  title,
  description,
  actions,
  className,
}: {
  eyebrow?: ReactNode;
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn("flex items-baseline justify-between gap-4 flex-wrap", className)}>
      <div>
        {eyebrow && (
          <span className="font-mono text-[11px] uppercase tracking-widest3 text-muted-foreground block mb-2">
            {eyebrow}
          </span>
        )}
        <h2 className="display-head text-3xl">{title}</h2>
        {description && <p className="text-sm text-muted-foreground mt-1">{description}</p>}
      </div>
      {actions && <div className="flex gap-2 items-center">{actions}</div>}
    </div>
  );
}
