"use client";

import Link from "next/link";
import { createColumnHelper } from "@tanstack/react-table";
import { Power, PowerOff, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { UserRoleBadge } from "@/components/admin/user-role-badge";
import { DataTableColumnHeader } from "@/components/admin/data-table/data-table-column-header";
import type { DataTableColumnDef, DataTableFeatures } from "@/components/admin/data-table/types";
import { formatDate } from "@/lib/utils";
import type { UserListItem } from "@/types/user.types";

export interface UserRowActions {
  onToggleStatus: (user: UserListItem) => void;
  onDelete: (user: UserListItem) => void;
  /** BR-001 usability affordance: hides the destructive actions on the
   * signed-in admin's own row. The server's 403 remains the real guard. */
  currentUserId: string | undefined;
}

const helper = createColumnHelper<DataTableFeatures, UserListItem>();

/** FR-001's column set. Role change (FR-005) is detail-only (SCR002) — not offered here. */
export function buildUserColumns(actions: UserRowActions): DataTableColumnDef<UserListItem>[] {
  return [
    helper.accessor("email", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Email" />,
      cell: ({ row }) => (
        <Button asChild variant="link" className="h-auto p-0 font-medium">
          <Link href={`/admin/users/${row.original.id}`}>{row.original.email}</Link>
        </Button>
      ),
    }),
    helper.accessor("full_name", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Full name" />,
    }),
    helper.accessor("role", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Role" />,
      cell: ({ getValue }) => <UserRoleBadge role={getValue()} />,
    }),
    helper.accessor("is_active", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Status" />,
      cell: ({ getValue }) => (getValue() ? <Badge>Active</Badge> : <Badge variant="outline">Inactive</Badge>),
    }),
    helper.accessor("created_at", {
      header: ({ column }) => <DataTableColumnHeader column={column} title="Created" />,
      cell: ({ getValue }) => <span className="text-sm text-muted-foreground">{formatDate(getValue())}</span>,
    }),
    helper.display({
      id: "actions",
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => {
        const u = row.original;
        if (actions.currentUserId === u.id) return null;
        return (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              aria-label={u.is_active ? `Deactivate ${u.email}` : `Activate ${u.email}`}
              onClick={() => actions.onToggleStatus(u)}
            >
              {u.is_active ? <PowerOff className="h-4 w-4" /> : <Power className="h-4 w-4" />}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-destructive"
              aria-label={`Delete ${u.email}`}
              onClick={() => actions.onDelete(u)}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        );
      },
    }),
  ];
}
