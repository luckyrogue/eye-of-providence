import type { ReactNode } from "react";
import { cn } from "@eop/ui";

export function IconButton({
  onClick,
  title,
  children,
  danger,
}: {
  onClick?: () => void;
  title?: string;
  children: ReactNode;
  danger?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      className={cn(
        "rounded-md p-1.5 text-muted-foreground transition-colors",
        danger
          ? "hover:text-destructive hover:bg-destructive/10"
          : "hover:text-foreground hover:bg-secondary",
      )}
    >
      {children}
    </button>
  );
}
