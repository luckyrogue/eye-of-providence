import { ChevronDown } from "lucide-react";
import { forwardRef, type SelectHTMLAttributes } from "react";
import { cn } from "./cn";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  mono?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ label, mono, className, children, ...props }, ref) => {
    const sel = (
      <div className={cn("relative inline-block", className)}>
        <select
          ref={ref}
          className={cn(
            "w-full appearance-none rounded-md border bg-background px-2 py-1.5 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-primary",
            mono && "font-mono",
          )}
          {...props}
        >
          {children}
        </select>
        <ChevronDown
          aria-hidden="true"
          className="pointer-events-none absolute right-2 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
        />
      </div>
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
