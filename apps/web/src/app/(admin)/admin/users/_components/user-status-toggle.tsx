"use client";

import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Switch } from "@/components/ui/switch";
import { useToggleUserStatus } from "@/hooks/use-user-mutations";
import { ApiError } from "@/lib/api/types";

interface UserStatusToggleProps {
  userId: string;
  isActive: boolean;
}

/** FR-004. BR-001 (self) and BR-002 (last admin) are server-side guards —
 * this control carries no client-side lockout logic; a 403/409 surfaces the
 * server's own message and the switch simply stays at its last known state. */
export function UserStatusToggle({ userId, isActive }: UserStatusToggleProps) {
  const toggleStatus = useToggleUserStatus();

  function handleChange(next: boolean) {
    toggleStatus.mutate(
      { id: userId, is_active: next },
      {
        onSuccess: () => toast.success(next ? "User activated" : "User deactivated"),
        onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not change this account's status."),
      },
    );
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-sm font-medium">Account active</span>
      <div className="flex items-center gap-2">
        {toggleStatus.isPending && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
        <Switch checked={isActive} disabled={toggleStatus.isPending} onCheckedChange={handleChange} aria-label="Toggle account active" />
      </div>
    </div>
  );
}
