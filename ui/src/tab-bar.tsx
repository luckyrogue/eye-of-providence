import type { ReactNode } from "react";
import { cn } from "./cn";

export function TabBar({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn("flex gap-1 text-sm flex-wrap", className)}>{children}</div>;
}

export function Tab({
  active,
  onClick,
  icon,
  children,
}: {
  active?: boolean;
  onClick?: () => void;
  icon?: ReactNode;
  children: ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex items-center gap-1.5 rounded-md px-2.5 py-1.5 transition-colors",
        active
          ? "bg-primary text-primary-foreground"
          : "text-muted-foreground hover:bg-secondary",
      )}
    >
      {icon}
      {children}
    </button>
  );
}
