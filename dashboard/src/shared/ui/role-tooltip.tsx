// Minimal tooltip wrapper для read-only состояний (observer role).
//
// Не вытягиваем @radix-ui/react-tooltip — для нашего use-case достаточно
// HTML `title` атрибута + aria-disabled и role="note". Каркас остаётся
// доступным без extra deps.
//
// Также экспортируем role helpers — единая точка для проверки role==='observer'.

import type { ReactNode } from "react";

export const READONLY_ROLES = new Set(["observer"]);

export function isReadOnlyRole(role: string | undefined | null): boolean {
  if (!role) return false;
  return READONLY_ROLES.has(role);
}

export function canMutate(role: string | undefined | null): boolean {
  return !isReadOnlyRole(role);
}

// RoleGate — скрывает children для observer role, опционально показывая
// disabled placeholder с tooltip-hint. Используется на mutation buttons.
export function RoleGate({
  role,
  children,
  fallback,
}: {
  role: string | undefined | null;
  children: ReactNode;
  fallback?: ReactNode;
}) {
  if (isReadOnlyRole(role)) return <>{fallback ?? null}</>;
  return <>{children}</>;
}

// ObserverHint — небольшой read-only badge с tooltip-text.
// Используется в местах, где надо явно сообщить юзеру про read-only.
export function ObserverHint({ label, hint }: { label: string; hint: string }) {
  return (
    <span
      role="note"
      title={hint}
      className="inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-[11px] text-muted-foreground"
      style={{ borderColor: "hsl(var(--border))" }}
    >
      <span
        className="h-1.5 w-1.5 rounded-full"
        style={{ background: "hsl(var(--muted-foreground))" }}
        aria-hidden
      />
      {label}
    </span>
  );
}
