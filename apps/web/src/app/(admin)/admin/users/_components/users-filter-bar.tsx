"use client";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { UserRole } from "@/types/user.types";

interface UsersFilterBarProps {
  role: UserRole | undefined;
  /** Raw URL value: "true" | "false" | undefined. */
  isActive: string | undefined;
  onRoleChange: (role: string | undefined) => void;
  onStatusChange: (isActive: string | undefined) => void;
}

/** FR-002: role + active-status filters, combinable with the search box. */
export function UsersFilterBar({ role, isActive, onRoleChange, onStatusChange }: UsersFilterBarProps) {
  return (
    <>
      <Select value={role ?? "all"} onValueChange={(v) => onRoleChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-40" aria-label="Role filter">
          <SelectValue placeholder="All roles" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All roles</SelectItem>
          <SelectItem value="admin">Admin</SelectItem>
          <SelectItem value="user">User</SelectItem>
        </SelectContent>
      </Select>

      <Select value={isActive ?? "all"} onValueChange={(v) => onStatusChange(v === "all" ? undefined : v)}>
        <SelectTrigger className="w-40" aria-label="Status filter">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All statuses</SelectItem>
          <SelectItem value="true">Active</SelectItem>
          <SelectItem value="false">Inactive</SelectItem>
        </SelectContent>
      </Select>
    </>
  );
}
