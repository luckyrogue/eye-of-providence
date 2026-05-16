// Готовая колонка-чекбокс для row selection (master в шапке + чекбокс в строке).
// API соответствует доке shadcn data-table.
import type { ColumnDef } from "@tanstack/react-table";

import { Checkbox } from "../checkbox";

export type SelectColumnLabels = {
  selectAll?: string;
  selectRow?: string;
};

export function selectColumn<TData>(labels?: SelectColumnLabels): ColumnDef<TData> {
  return {
    id: "__select__",
    enableSorting: false,
    enableHiding: false,
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label={labels?.selectAll ?? "Select all"}
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label={labels?.selectRow ?? "Select row"}
      />
    ),
    size: 32,
  };
}
