"use client";

import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import { UserRoleBadge } from "@/components/admin/user-role-badge";
import { formatDate } from "@/lib/utils";
import type { UserDetail } from "@/types/user.types";

interface UserProfileCardProps {
  user: UserDetail;
}

/** FR-003: read-only profile plus the two lifetime history counts. Never
 * renders password_hash — the field does not exist on UserDetail. */
export function UserProfileCard({ user }: UserProfileCardProps) {
  return (
    <section className="rounded-lg border p-5">
      <h2 className="text-sm font-semibold mb-4">Profile</h2>
      <dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
        <Field label="Full name" value={user.full_name} />
        <Field label="Email" value={user.email} />
        <Field label="Phone" value={user.phone || "—"} />
        <Field label="Role" value={<UserRoleBadge role={user.role} />} />
        <Field label="Status" value={user.is_active ? <Badge>Active</Badge> : <Badge variant="outline">Inactive</Badge>} />
        <Field label="Joined" value={formatDate(user.created_at)} />
      </dl>

      <div className="mt-5 flex gap-3">
        <Badge variant="secondary" className="text-sm px-3 py-1">
          {user.booking_count} booking{user.booking_count === 1 ? "" : "s"}
        </Badge>
        <Badge variant="secondary" className="text-sm px-3 py-1">
          {user.review_count} review{user.review_count === 1 ? "" : "s"}
        </Badge>
      </div>
    </section>
  );
}

function Field({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 break-words">{value}</dd>
    </div>
  );
}
