"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/admin/page-header";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { UserRoleBadge } from "@/components/admin/user-role-badge";
import { useAdminSession } from "@/hooks/use-admin-session";
import { useDeleteUser } from "@/hooks/use-user-mutations";
import { getUser } from "@/lib/api/users";
import { userKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";
import { UserProfileCard } from "./user-profile-card";
import { UserRoleDialog } from "./user-role-dialog";
import { UserStatusToggle } from "./user-status-toggle";

interface UserDetailClientProps {
  id: string;
}

/**
 * SCR002_UserDetailScreen (FR-003…006). Read-only profile apart from the
 * status toggle, role dialog and delete. Self-target actions are hidden
 * (BR-001 usability affordance) — the server's 403 is the real guard.
 */
export function UserDetailClient({ id }: UserDetailClientProps) {
  const router = useRouter();
  const { data: session } = useAdminSession();
  const [roleDialogOpen, setRoleDialogOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  const { data: user, isLoading, isError, refetch } = useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => getUser(id),
  });

  const remove = useDeleteUser();

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-9 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (isError || !user) {
    return (
      <div className="rounded-lg border border-destructive/30 bg-destructive/5 p-6 text-sm flex items-center justify-between">
        <span>Could not load this user.</span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Retry
          </Button>
          <Button asChild variant="ghost" size="sm">
            <Link href="/admin/users">Back to list</Link>
          </Button>
        </div>
      </div>
    );
  }

  const isSelf = session?.id === user.id;
  const userId = user.id;

  function handleDelete() {
    remove.mutate(userId, {
      onSuccess: () => {
        toast.success("User deleted");
        router.push("/admin/users");
      },
      // BR-003's 409 names the blocking booking count — surfaced verbatim,
      // dialog stays open so the admin can read it and cancel.
      onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not delete this user."),
    });
  }

  return (
    <>
      <Button asChild variant="ghost" size="sm" className="mb-2 -ml-2">
        <Link href="/admin/users">
          <ArrowLeft className="h-4 w-4 mr-2" />
          User Management
        </Link>
      </Button>

      <PageHeader title={user.full_name} description={user.email}>
        <UserRoleBadge role={user.role} />
      </PageHeader>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <UserProfileCard user={user} />
        </div>

        <div className="space-y-6">
          <section className="rounded-lg border p-5 space-y-4">
            <h2 className="text-sm font-semibold">Actions</h2>
            {isSelf ? (
              <p className="text-sm text-muted-foreground">You can&apos;t change your own account from here.</p>
            ) : (
              <>
                <UserStatusToggle userId={user.id} isActive={user.is_active} />
                <Button variant="outline" className="w-full" onClick={() => setRoleDialogOpen(true)}>
                  Change role
                </Button>
                <Button variant="destructive" className="w-full" onClick={() => setDeleteOpen(true)}>
                  Delete user
                </Button>
              </>
            )}
          </section>
        </div>
      </div>

      {!isSelf && <UserRoleDialog user={user} open={roleDialogOpen} onOpenChange={setRoleDialogOpen} />}

      {!isSelf && (
        <ConfirmDialog
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          title="Delete user"
          destructive
          confirmLabel="Delete"
          isPending={remove.isPending}
          onConfirm={handleDelete}
          description={
            <p>
              This will remove <strong>{user.email}</strong> from the active listing. Their bookings and reviews remain untouched.
            </p>
          }
        />
      )}
    </>
  );
}
