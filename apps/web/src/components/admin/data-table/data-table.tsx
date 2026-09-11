"use client";

import { functionalUpdate, useTable, type PaginationState, type RowData, type SortingState } from "@tanstack/react-table";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { DataTablePagination } from "./data-table-pagination";
import { dataTableFeatures, type DataTableColumnDef, type DataTablePagingState, type DataTableSortState } from "./types";

interface DataTableProps<TData extends RowData> extends DataTablePagingState, DataTableSortState {
  columns: DataTableColumnDef<TData>[];
  data: TData[];
  getRowId: (row: TData) => string;
  isLoading?: boolean;
  emptyMessage?: string;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onSortChange: (sort: DataTableSortState) => void;
}

const SKELETON_ROWS = 5;

/**
 * Generic server-driven table (react-table v9, manual sorting + pagination).
 * The URL/query hook owns the state; this component only renders it and
 * reports the admin's intent back through the on* callbacks.
 */
export function DataTable<TData extends RowData>({
  columns,
  data,
  getRowId,
  isLoading = false,
  emptyMessage = "No results found.",
  page,
  pageSize,
  total,
  sortBy,
  sortDir,
  onPageChange,
  onPageSizeChange,
  onSortChange,
}: DataTableProps<TData>) {
  const sorting: SortingState = sortBy ? [{ id: sortBy, desc: sortDir === "desc" }] : [];
  const pagination: PaginationState = { pageIndex: page - 1, pageSize };

  const table = useTable({
    features: dataTableFeatures,
    columns,
    data,
    getRowId,
    manualSorting: true,
    manualPagination: true,
    enableMultiSort: false,
    rowCount: total,
    state: { sorting, pagination },
    onSortingChange: (updater) => {
      const next = functionalUpdate(updater, sorting);
      const first = next[0];
      onSortChange(first ? { sortBy: first.id, sortDir: first.desc ? "desc" : "asc" } : {});
    },
    onPaginationChange: (updater) => {
      const next = functionalUpdate(updater, pagination);
      if (next.pageSize !== pageSize) onPageSizeChange(next.pageSize);
      else onPageChange(next.pageIndex + 1);
    },
  });

  const rows = table.getRowModel().rows;
  const columnCount = columns.length;

  return (
    <div className="rounded-lg border bg-card">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id} colSpan={header.colSpan}>
                  {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {isLoading ? (
            Array.from({ length: SKELETON_ROWS }).map((_, i) => (
              <TableRow key={`skeleton-${i}`}>
                <TableCell colSpan={columnCount}>
                  <Skeleton className="h-6 w-full" />
                </TableCell>
              </TableRow>
            ))
          ) : rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={columnCount} className="h-24 text-center text-muted-foreground">
                {emptyMessage}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((row) => (
              <TableRow key={row.id}>
                {row.getAllCells().map((cell) => (
                  <TableCell key={cell.id}>
                    <table.FlexRender cell={cell} />
                  </TableCell>
                ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
      <DataTablePagination page={page} pageSize={pageSize} total={total} onPageChange={onPageChange} onPageSizeChange={onPageSizeChange} />
    </div>
  );
}
