import { forwardRef, useId, type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "../../lib/cn";
export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  label?: ReactNode;
  labelClassName?: string;
};
export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, label, labelClassName, id, ...props }, ref) => {
    const uid = useId();
    const inputId = label ? (id ?? uid) : id;
    const control = (
      <input
        ref={ref}
        id={inputId}
        type={type}
        className={cn(
          "box-border h-10 w-full rounded-md border bg-background px-3 py-2 text-sm leading-5",
          "ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium",
          "placeholder:text-muted-foreground",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2",
          "disabled:cursor-not-allowed disabled:opacity-50",
          className,
        )}
        {...props}
      />
    );
    if (!label) {
      return control;
    }
    return (
      <div className="w-full">
        <label
          htmlFor={inputId}
          className={cn("mb-1 block text-xs text-muted-foreground", labelClassName)}
        >
          {label}
        </label>
        {control}
      </div>
    );
  },
);
Input.displayName = "Input";
