import { Badge } from "@/components/ui/badge";
import type { UserRole } from "@/types/user.types";

interface UserRoleBadgeProps {
  role: UserRole;
  className?: string;
}

/** Shared admin/user badge for the users list and detail screens. */
export function UserRoleBadge({ role, className }: UserRoleBadgeProps) {
  return (
    <Badge variant={role === "admin" ? "default" : "outline"} className={className}>
      {role === "admin" ? "Admin" : "User"}
    </Badge>
  );
}
