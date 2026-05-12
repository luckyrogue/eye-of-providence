// Sortable header для DataTable: рендерит лейбл + иконку текущего направления
// сортировки. Если у column.getCanSort() == false — рендерит plain title.
import type { Column } from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ChevronsUpDown } from "lucide-react";

import { cn } from "../../lib/cn";
import { Button } from "../button";

export type DataTableColumnHeaderProps<TData, TValue> = {
  column: Column<TData, TValue>;
  title: React.ReactNode;
  className?: string;
  align?: "left" | "right" | "center";
};

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  className,
  align = "left",
}: DataTableColumnHeaderProps<TData, TValue>) {
  const alignClass =
    align === "right" ? "justify-end" : align === "center" ? "justify-center" : "justify-start";

  if (!column.getCanSort()) {
    return <div className={cn("flex items-center gap-1", alignClass, className)}>{title}</div>;
  }

  const sorted = column.getIsSorted();

  return (
    <div className={cn("flex items-center", alignClass, className)}>
      <Button
        variant="ghost"
        size="sm"
        className="-ml-2 h-10 px-2 text-xs uppercase tracking-wide text-muted-foreground hover:text-foreground"
        onClick={() => column.toggleSorting(sorted === "asc")}
      >
        <span>{title}</span>
        {sorted === "desc" ? (
          <ArrowDown className="ml-1 h-3 w-3" />
        ) : sorted === "asc" ? (
          <ArrowUp className="ml-1 h-3 w-3" />
        ) : (
          <ChevronsUpDown className="ml-1 h-3 w-3 opacity-50" />
        )}
      </Button>
    </div>
  );
}
