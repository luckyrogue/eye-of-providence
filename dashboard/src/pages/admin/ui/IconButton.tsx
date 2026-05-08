import type { ReactNode } from "react";
import { cn } from "@eop/ui";

// Иконочная кнопка. `title` обязателен — он же используется как aria-label,
// чтобы скрин-ридер озвучивал назначение (иначе кнопка читается как пустая).
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
        "rounded-md p-1.5 text-muted-foreground transition-colors disabled:opacity-50",
        danger
          ? "hover:text-destructive hover:bg-destructive/10"
          : "hover:text-foreground hover:bg-secondary",
      )}
    >
      {children}
    </button>
  );
}
