import { forwardRef, type SelectHTMLAttributes } from "react";
import { cn } from "./cn";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  mono?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, mono, className, children, ...props }, ref) => {
    const sel = (
      <select
        ref={ref}
        className={cn(
          "rounded-md border bg-background px-2 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary",
          mono && "font-mono",
          className,
        )}
        {...props}
      >
        {children}
      </select>
    );
    if (!label) return sel;
    return (
      <div className="space-y-1">
        <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
          {label}
        </label>
        {sel}
      </div>
    );
  },
);
Select.displayName = "Select";
