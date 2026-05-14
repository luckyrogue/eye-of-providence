import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { cn } from "@eop/ui";
export function NavItem({
  to,
  icon,
  children,
  accent,
}: {
  to: string;
  icon: ReactNode;
  children: ReactNode;
  accent?: boolean;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-1.5 rounded-md px-3 py-1.5 transition-colors",
          isActive
            ? "bg-primary text-primary-foreground"
            : accent
              ? "text-amber-600 dark:text-amber-400 hover:bg-amber-500/10"
              : "text-muted-foreground hover:bg-secondary",
        )
      }
    >
      {icon} {children}
    </NavLink>
  );
}
