"use client";

import { useMemo, useState } from "react";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageHeader } from "@/components/admin/page-header";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableToolbar } from "@/components/admin/data-table/data-table-toolbar";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { createCategory, deleteCategory, listCategories, reorderCategory, updateCategory } from "@/lib/api/categories";
import { categoryKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import type { CategoryFormValues } from "@/lib/validation/category-schema";
import type { CategoryListItem } from "@/types/category.types";
import { buildCategoryColumns } from "./categories-columns";
import { CategoryDeleteDialog } from "./category-delete-dialog";
import { CategoryFormDialog } from "./category-form-dialog";

function toPayload(v: CategoryFormValues) {
  return { name: v.name, slug: v.slug, description: v.description || undefined, image_url: v.image_url || undefined, sort_order: v.sort_order, is_active: v.is_active };
}

export function CategoriesClient() {
  const list = useListQueryState({ filterKeys: ["is_active"] });
  const queryClient = useQueryClient();
  const [formTarget, setFormTarget] = useState<"create" | CategoryListItem | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<CategoryListItem | null>(null);

  const isActive = list.filters.is_active;
  const apiQuery = { ...list.query, is_active: isActive === undefined ? undefined : isActive === "true" };
  const queryKey = categoryKeys.list(apiQuery);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey,
    queryFn: () => listCategories(apiQuery),
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  const invalidate = () => queryClient.invalidateQueries({ queryKey: categoryKeys.lists() });
  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Request failed");

  const reorder = useMutation({
    mutationFn: ({ id, target }: { id: string; target: number }) => reorderCategory(id, target),
    // The server returns the full ordered set but without tour_count; a row
    // crossing a page boundary would show a wrong count if we patched the
    // cache from it, so refetch instead (keepPreviousData keeps rows visible).
    onSuccess: () => invalidate(),
    onError,
  });
  const remove = useMutation({
    mutationFn: (c: CategoryListItem) => deleteCategory(c.id),
    onSuccess: () => { toast.success("Category deleted"); setDeleteTarget(null); void invalidate(); },
    onError,
  });

  // Up/down is only meaningful in the default display order with no filters.
  const canReorder = !list.sortBy && !list.search && isActive === undefined;
  const rankOf = (c: CategoryListItem) => (list.page - 1) * list.pageSize + items.findIndex((x) => x.id === c.id);

  const columns = useMemo(
    () =>
      buildCategoryColumns({
        onEdit: setFormTarget,
        onDelete: setDeleteTarget,
        onMove: (c, dir) => reorder.mutate({ id: c.id, target: rankOf(c) + dir }),
        canReorder,
        isFirst: (c) => rankOf(c) === 0,
        isLast: (c) => rankOf(c) === total - 1,
      }),
    // eslint-disable-next-line react-hooks/exhaustive-deps -- rankOf closes over items/paging which are covered below
    [items, total, list.page, list.pageSize, canReorder, reorder.mutate],
  );

  async function submitForm(values: CategoryFormValues) {
    if (formTarget && formTarget !== "create") await updateCategory(formTarget.id, toPayload(values));
    else await createCategory(toPayload(values));
    await invalidate();
  }

  return (
    <>
      <PageHeader title="Tour Categories" description="Group tour packages and control their display order.">
        <Button onClick={() => setFormTarget("create")}><Plus className="h-4 w-4 mr-2" />New Category</Button>
      </PageHeader>

      <DataTableToolbar search={list.search} onSearchChange={(search) => list.set({ search })} placeholder="Search by name…">
        <Select value={isActive ?? "all"} onValueChange={(v) => list.set({ filters: { is_active: v === "all" ? undefined : v } })}>
          <SelectTrigger className="w-40" aria-label="Status filter"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            <SelectItem value="true">Active</SelectItem>
            <SelectItem value="false">Inactive</SelectItem>
          </SelectContent>
        </Select>
      </DataTableToolbar>

      {isError ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
          <span>Could not load categories.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>Retry</Button>
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          getRowId={(c) => c.id}
          isLoading={isLoading}
          emptyMessage="No categories found."
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

      <CategoryFormDialog target={formTarget} onOpenChange={(o) => !o && setFormTarget(null)} onSubmit={submitForm} />
      <CategoryDeleteDialog category={deleteTarget} onOpenChange={(o) => !o && setDeleteTarget(null)} onConfirm={(c) => remove.mutate(c)} isPending={remove.isPending} />
    </>
  );
}
