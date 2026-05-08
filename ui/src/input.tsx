import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "./cn";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: ReactNode;
  hint?: ReactNode;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, hint, error, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label className="font-mono text-[10px] uppercase tracking-widest2 text-muted-foreground">
            {label}
          </label>
        )}
        <input
          ref={ref}
          className={cn(
            "w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary",
            error && "border-destructive focus:ring-destructive",
            className,
          )}
          {...props}
        />
        {error && <div className="text-xs text-destructive">{error}</div>}
        {hint && !error && <div className="text-xs text-muted-foreground">{hint}</div>}
      </div>
    );
  },
);
Input.displayName = "Input";
