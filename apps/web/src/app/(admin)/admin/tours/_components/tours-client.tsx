"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableToolbar } from "@/components/admin/data-table/data-table-toolbar";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { useDeleteTour, useUpdateTourStatus } from "@/hooks/use-tour-mutations";
import { listTours, type TourListQuery } from "@/lib/api/tours";
import { listCategories, type CategoryListQuery } from "@/lib/api/categories";
import { categoryKeys, tourKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import type { TourListItem, TourStatus } from "@/types/tour.types";
import { buildTourColumns } from "./tours-columns";
import { ToursFilterBar } from "./tours-filter-bar";
import { TourDeleteDialog } from "./tour-delete-dialog";

const FILTER_KEYS = ["category_id", "status", "price_min", "price_max"];

export function ToursClient() {
  const list = useListQueryState({ filterKeys: FILTER_KEYS });
  const [deleteTarget, setDeleteTarget] = useState<TourListItem | null>(null);

  const apiQuery: TourListQuery = {
    ...list.query,
    category_id: list.filters.category_id,
    status: list.filters.status,
    price_min: list.filters.price_min ? Number(list.filters.price_min) : undefined,
    price_max: list.filters.price_max ? Number(list.filters.price_max) : undefined,
  };

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: tourKeys.list(apiQuery),
    queryFn: () => listTours(apiQuery),
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  // Category select + name lookup for the Title column (List never joins
  // category — tours-columns.tsx's comment).
  const activeCategoryQuery: CategoryListQuery = { page: 1, page_size: 100, is_active: true };
  const { data: categoryPage } = useQuery({
    queryKey: categoryKeys.list(activeCategoryQuery),
    queryFn: () => listCategories(activeCategoryQuery),
  });
  const categories = categoryPage?.items ?? [];
  const categoryNameById = useMemo(() => new Map((categoryPage?.items ?? []).map((c) => [c.id, c.name])), [categoryPage]);

  const updateStatus = useUpdateTourStatus();
  const remove = useDeleteTour();
  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Request failed");

  function changeStatus(tour: TourListItem, status: TourStatus) {
    updateStatus.mutate({ id: tour.id, status }, { onError });
  }

  const columns = useMemo(
    () =>
      buildTourColumns({
        categoryNameById,
        onPublish: (t) => changeStatus(t, "published"),
        onArchive: (t) => changeStatus(t, "archived"),
        onReactivate: (t) => changeStatus(t, "published"),
        onDelete: setDeleteTarget,
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- changeStatus closes over updateStatus.mutate, which is stable across renders
    [categoryNameById],
  );

  return (
    <>
      <PageHeader title="Tour Packages" description="Create, publish and manage the tour catalogue.">
        <Button asChild>
          <Link href="/admin/tours/new">
            <Plus className="h-4 w-4 mr-2" />
            New Tour
          </Link>
        </Button>
      </PageHeader>

      <DataTableToolbar search={list.search} onSearchChange={(search) => list.set({ search })} placeholder="Search by title or destination…">
        <ToursFilterBar
          categories={categories}
          categoryId={list.filters.category_id}
          status={list.filters.status as TourStatus | undefined}
          priceMin={list.filters.price_min}
          priceMax={list.filters.price_max}
          onCategoryChange={(v) => list.set({ filters: { category_id: v } })}
          onStatusChange={(v) => list.set({ filters: { status: v } })}
          onPriceChange={(bounds) => list.set({ filters: bounds })}
        />
      </DataTableToolbar>

      {isError ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
          <span>Could not load tours.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          getRowId={(t) => t.id}
          isLoading={isLoading}
          emptyMessage="No tours found."
          page={list.page}
          pageSize={list.pageSize}
          total={total}
          sortBy={list.sortBy}
          sortDir={list.sortDir}
          onPageChange={(page) => list.set({ page })}
          onPageSizeChange={(pageSize) => list.set({ pageSize })}
          onSortChange={({ sortBy, sortDir }) => list.set({ sortBy, sortDir })}
        />
      )}

      <TourDeleteDialog
        tour={deleteTarget}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        onDeleted={() => setDeleteTarget(null)}
        deleteMutation={remove}
      />
    </>
  );
}
