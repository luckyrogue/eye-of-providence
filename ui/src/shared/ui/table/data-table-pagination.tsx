import type { Table } from "@tanstack/react-table";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { Button } from "../button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../select";
import type { DataTableLabels } from "./types";
export type DataTablePaginationProps<TData> = {
  table: Table<TData>;
  pageSizeOptions?: number[];
  showSelectionCount?: boolean;
  labels?: DataTableLabels;
};
const DEFAULT_PAGE_SIZE_OPTIONS = [10, 25, 50, 100];
export function DataTablePagination<TData>({
  table,
  pageSizeOptions = DEFAULT_PAGE_SIZE_OPTIONS,
  showSelectionCount = false,
  labels,
}: DataTablePaginationProps<TData>) {
  const selected = table.getFilteredSelectedRowModel().rows.length;
  const total = table.getFilteredRowModel().rows.length;
  const pageIndex = table.getState().pagination.pageIndex;
  const pageCount = table.getPageCount();
  return (
    <div className="flex items-center justify-between gap-2 px-1 py-2">
      <div className="flex-1 text-xs text-muted-foreground">
        {showSelectionCount
          ? labels?.selectedOfTotal
            ? labels.selectedOfTotal(selected, total)
            : `${selected} of ${total} row(s) selected.`
          : null}
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <p className="text-xs font-medium text-muted-foreground">
            {labels?.rowsPerPage ?? "Rows per page"}
          </p>
          <Select
            value={`${table.getState().pagination.pageSize}`}
            onValueChange={(v) => table.setPageSize(Number(v))}
          >
            <SelectTrigger className="h-10 w-[4.5rem] text-xs">
              <SelectValue placeholder={`${table.getState().pagination.pageSize}`} />
            </SelectTrigger>
            <SelectContent side="top">
              {pageSizeOptions.map((s) => (
                <SelectItem key={s} value={`${s}`} className="text-xs">
                  {s}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex w-[7rem] items-center justify-center text-xs font-medium text-muted-foreground">
          {labels?.pageOfTotal
            ? labels.pageOfTotal(pageIndex + 1, Math.max(pageCount, 1))
            : `Page ${pageIndex + 1} of ${Math.max(pageCount, 1)}`}
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            onClick={() => table.setPageIndex(0)}
            disabled={!table.getCanPreviousPage()}
            aria-label={labels?.firstPage ?? "First page"}
          >
            <ChevronsLeft className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
            aria-label={labels?.previous ?? "Previous"}
          >
            <ChevronLeft className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
            aria-label={labels?.next ?? "Next"}
          >
            <ChevronRight className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            onClick={() => table.setPageIndex(pageCount - 1)}
            disabled={!table.getCanNextPage()}
            aria-label={labels?.lastPage ?? "Last page"}
          >
            <ChevronsRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
}
