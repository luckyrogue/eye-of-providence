// Универсальный slot ⋯ row-actions для DataTable.
// Caller передаёт DropdownMenuItem-узлы как children.
import { MoreHorizontal } from "lucide-react";

import { Button } from "../button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "../dropdown-menu";

export type DataTableRowActionsProps = {
  children: React.ReactNode;
  align?: "start" | "center" | "end";
  triggerLabel?: string;
  contentClassName?: string;
};

export function DataTableRowActions({
  children,
  align = "end",
  triggerLabel = "Open menu",
  contentClassName = "w-44",
}: DataTableRowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-7 w-7" aria-label={triggerLabel}>
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align} className={contentClassName}>
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
