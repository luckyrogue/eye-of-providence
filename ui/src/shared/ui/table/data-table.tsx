// <DataTable<TData,TValue>> — Reusable Component из shadcn доки поверх @tanstack/react-table.
// Покрывает Basic Table, Row Actions, Pagination, Sorting, Filtering, Visibility, Row Selection.
// i18n-агностично: caller передаёт labels, headers/cells рендерит сам.
import * as React from "react";
import {
  type ColumnDef,
  type ColumnFiltersState,
  type Row,
  type RowSelectionState,
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";

import { cn } from "../../lib/cn";
import { Input } from "../input";
import { Skeleton } from "../skeleton";
import { DataTablePagination } from "./data-table-pagination";
import { DataTableViewOptions } from "./data-table-view-options";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./table";
import type { DataTableLabels } from "./types";

export type DataTableProps<TData, TValue> = {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];

  // Filtering — встроенный input, фильтрует по одной колонке.
  filterableColumn?: { id: string; placeholder?: string };

  // Visibility
  enableColumnVisibility?: boolean;
  initialColumnVisibility?: VisibilityState;

  // Pagination
  enablePagination?: boolean;
  pageSize?: number;
  pageSizeOptions?: number[];

  // Row selection
  enableRowSelection?: boolean;
  onSelectionChange?: (selectedRows: TData[]) => void;

  // Expandable rows (например, AddMemberRow в TeamsTable).
  getRowCanExpand?: (row: Row<TData>) => boolean;
  renderSubComponent?: (row: Row<TData>) => React.ReactNode;

  // Toolbar slot для caller-кнопок (рядом с фильтром / Columns).
  toolbarRight?: React.ReactNode;

  // States
  isLoading?: boolean;
  emptyState?: React.ReactNode;

  // i18n-агностично
  labels?: DataTableLabels;

  className?: string;
  tableClassName?: string;
};

export function DataTable<TData, TValue>({
  columns,
  data,
  filterableColumn,
  enableColumnVisibility = false,
  initialColumnVisibility,
  enablePagination = false,
  pageSize,
  pageSizeOptions,
  enableRowSelection = false,
  onSelectionChange,
  getRowCanExpand,
  renderSubComponent,
  toolbarRight,
  isLoading = false,
  emptyState,
  labels,
  className,
  tableClassName,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>(
    initialColumnVisibility ?? {},
  );
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
    enableRowSelection,
    getRowCanExpand,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: enablePagination ? getPaginationRowModel() : undefined,
    getExpandedRowModel: getRowCanExpand ? getExpandedRowModel() : undefined,
    initialState: {
      pagination: pageSize ? { pageIndex: 0, pageSize } : undefined,
    },
  });

  // Прокидываем выделенные строки наверх (TanStack хранит их как { [rowId]: true }).
  React.useEffect(() => {
    if (!onSelectionChange) return;
    const rows = table.getFilteredSelectedRowModel().rows.map((r) => r.original);
    onSelectionChange(rows);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rowSelection]);

  const filterColumn = filterableColumn ? table.getColumn(filterableColumn.id) : undefined;

  const showToolbar = !!filterableColumn || enableColumnVisibility || !!toolbarRight;

  return (
    <div className={cn("w-full space-y-2", className)}>
      {showToolbar && (
        <div className="flex items-center gap-2 py-1">
          {filterColumn && (
            <Input
              placeholder={
                filterableColumn?.placeholder ?? labels?.filterPlaceholder ?? "Filter..."
              }
              value={(filterColumn.getFilterValue() as string) ?? ""}
              onChange={(e) => filterColumn.setFilterValue(e.target.value)}
              className="h-8 max-w-xs text-sm"
            />
          )}
          <div className="ml-auto flex items-center gap-2">
            {toolbarRight}
            {enableColumnVisibility && <DataTableViewOptions table={table} labels={labels} />}
          </div>
        </div>
      )}

      <div className={cn("overflow-hidden rounded-md border", tableClassName)}>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} style={{ width: header.getSize() || undefined }}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: pageSize ?? 5 }).map((_, i) => (
                <TableRow key={`__skeleton_${i}`}>
                  {table.getVisibleLeafColumns().map((col) => (
                    <TableCell key={col.id}>
                      <Skeleton className="h-4 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : table.getRowModel().rows.length > 0 ? (
              table.getRowModel().rows.map((row) => (
                <React.Fragment key={row.id}>
                  <TableRow data-state={row.getIsSelected() ? "selected" : undefined}>
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                  {row.getIsExpanded() && renderSubComponent ? (
                    <TableRow data-state="expanded">
                      <TableCell colSpan={row.getVisibleCells().length} className="bg-muted/20 p-0">
                        {renderSubComponent(row)}
                      </TableCell>
                    </TableRow>
                  ) : null}
                </React.Fragment>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={table.getVisibleLeafColumns().length}
                  className="h-24 text-center text-sm text-muted-foreground"
                >
                  {emptyState ?? labels?.noResults ?? "No results."}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {enablePagination && (
        <DataTablePagination
          table={table}
          pageSizeOptions={pageSizeOptions}
          showSelectionCount={enableRowSelection}
          labels={labels}
        />
      )}
    </div>
  );
}
