import { rowPaginationFeature, rowSortingFeature, tableFeatures, type ColumnDef, type RowData } from "@tanstack/react-table";

/**
 * react-table v9 registers features explicitly. Only the two the admin
 * tables use are registered (phase-03: "register only the features used").
 * The Go API drives both, so every table sets manualSorting/manualPagination.
 */
export const dataTableFeatures = tableFeatures({ rowSortingFeature, rowPaginationFeature });

export type DataTableFeatures = typeof dataTableFeatures;

/**
 * Column definition type every admin screen's `*-columns.tsx` exports.
 * TValue is `any` on purpose: v9's ColumnDef is invariant in its value type,
 * so a `ColumnDef<…, number>` from `helper.accessor("sort_order", …)` would
 * not fit an `unknown` slot. This mirrors what v9's own `columnHelper.columns`
 * returns.
 */
export type DataTableColumnDef<TData extends RowData> = ColumnDef<DataTableFeatures, TData, any>;

export type SortDir = "asc" | "desc";

export interface DataTableSortState {
  sortBy?: string;
  sortDir?: SortDir;
}

export interface DataTablePagingState {
  /** 1-based, matching the API and the URL. */
  page: number;
  pageSize: number;
  total: number;
}
