import type { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { cn } from "@eop/ui";
export function BottomNavItem({
  to,
  icon,
  label,
  accent,
}: {
  to: string;
  icon: ReactNode;
  label: string;
  accent?: boolean;
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        cn(
          "relative flex-1 flex flex-col items-center justify-center gap-0.5 py-2 min-h-[56px] transition-colors",
          isActive
            ? accent
              ? "text-amber-600 dark:text-amber-400"
              : "text-primary"
            : "text-muted-foreground hover:text-foreground active:bg-secondary/30",
        )
      }
    >
      {({ isActive }) => (
        <>
          <span
            className={cn(
              "absolute top-0 left-1/2 -translate-x-1/2 h-0.5 w-8 rounded-full transition-all",
              isActive ? (accent ? "bg-amber-500" : "bg-primary") : "bg-transparent",
            )}
          />
          {icon}
          <span className="text-[10px] font-medium leading-none">{label}</span>
        </>
      )}
    </NavLink>
  );
}
