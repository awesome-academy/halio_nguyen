"use client";

import { useMemo, useState } from "react";
import { keepPreviousData, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/admin/page-header";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { DataTable } from "@/components/admin/data-table/data-table";
import { DataTableToolbar } from "@/components/admin/data-table/data-table-toolbar";
import { useListQueryState } from "@/hooks/use-list-query-state";
import { useAdminSession } from "@/hooks/use-admin-session";
import { useDeleteUser, useToggleUserStatus } from "@/hooks/use-user-mutations";
import { listUsers, type UserListQuery } from "@/lib/api/users";
import { userKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import type { UserListItem, UserRole } from "@/types/user.types";
import { buildUserColumns } from "./users-columns";
import { UsersFilterBar } from "./users-filter-bar";

const FILTER_KEYS = ["role", "is_active"];

/** SCR001_UserListScreen. Accounts are never created here — role-promotion
 * (detail-only, FR-005) is the sanctioned way to mint an admin (D1); this
 * screen only reads, toggles status, and deletes. */
export function UsersClient() {
  const list = useListQueryState({ filterKeys: FILTER_KEYS });
  const queryClient = useQueryClient();
  const { data: session } = useAdminSession();
  const [statusTarget, setStatusTarget] = useState<UserListItem | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<UserListItem | null>(null);

  const role = list.filters.role as UserRole | undefined;
  const isActive = list.filters.is_active;
  const apiQuery: UserListQuery = { ...list.query, role, is_active: isActive === undefined ? undefined : isActive === "true" };

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: userKeys.list(apiQuery),
    queryFn: () => listUsers(apiQuery),
    placeholderData: keepPreviousData,
  });
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  const toggleStatus = useToggleUserStatus();
  const remove = useDeleteUser();

  // BR-002's 409 message *is* the recovery instruction (promote a second
  // admin first) and BR-003's names the blocking booking count — both are
  // rendered verbatim, never replaced with a generic message.
  const onError = (err: unknown) => toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");

  const columns = useMemo(
    () => buildUserColumns({ onToggleStatus: setStatusTarget, onDelete: setDeleteTarget, currentUserId: session?.id }),
    [session?.id],
  );

  function confirmToggle() {
    if (!statusTarget) return;
    toggleStatus.mutate(
      { id: statusTarget.id, is_active: !statusTarget.is_active },
      {
        onSuccess: () => {
          toast.success(statusTarget.is_active ? "User deactivated" : "User activated");
          setStatusTarget(null);
        },
        onError,
      },
    );
  }

  function confirmDelete() {
    if (!deleteTarget) return;
    remove.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("User deleted");
        setDeleteTarget(null);
      },
      onError,
    });
  }

  return (
    <>
      <PageHeader title="User Management" description="Browse, search and moderate registered platform accounts." />

      <DataTableToolbar search={list.search} onSearchChange={(search) => list.set({ search })} placeholder="Search by email or name…">
        <UsersFilterBar
          role={role}
          isActive={isActive}
          onRoleChange={(v) => list.set({ filters: { role: v } })}
          onStatusChange={(v) => list.set({ filters: { is_active: v } })}
        />
      </DataTableToolbar>

      {isError ? (
        <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
          <span>Could not load users.</span>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={items}
          getRowId={(u) => u.id}
          isLoading={isLoading}
          emptyMessage="No users match your search."
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

      <ConfirmDialog
        open={statusTarget !== null}
        onOpenChange={(o) => !o && setStatusTarget(null)}
        title={statusTarget?.is_active ? "Deactivate user" : "Activate user"}
        destructive={statusTarget?.is_active}
        confirmLabel={statusTarget?.is_active ? "Deactivate" : "Activate"}
        isPending={toggleStatus.isPending}
        onConfirm={confirmToggle}
        description={
          statusTarget && (
            <p>
              {statusTarget.is_active ? "Deactivating" : "Activating"} <strong>{statusTarget.email}</strong>{" "}
              {statusTarget.is_active ? "revokes their ability to sign in." : "restores their ability to sign in."}
            </p>
          )
        }
      />

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => !o && setDeleteTarget(null)}
        title="Delete user"
        destructive
        confirmLabel="Delete"
        isPending={remove.isPending}
        onConfirm={confirmDelete}
        description={
          deleteTarget && (
            <p>
              This will remove <strong>{deleteTarget.email}</strong> from the active listing. Their bookings and reviews remain untouched.
            </p>
          )
        }
      />
    </>
  );
}
