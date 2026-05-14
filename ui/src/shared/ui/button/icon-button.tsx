import type { ReactNode } from "react";
import { cn } from "../../lib/cn";
export function IconButton({
  onClick,
  title,
  children,
  danger,
  disabled,
}: {
  onClick?: () => void;
  title: string;
  children: ReactNode;
  danger?: boolean;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={title}
      aria-label={title}
      disabled={disabled}
      className={cn(
        "inline-flex h-10 w-10 items-center justify-center rounded-md text-muted-foreground transition-colors disabled:opacity-50",
        danger
          ? "hover:text-destructive hover:bg-destructive/10"
          : "hover:text-foreground hover:bg-secondary",
      )}
    >
      {children}
    </button>
  );
}
