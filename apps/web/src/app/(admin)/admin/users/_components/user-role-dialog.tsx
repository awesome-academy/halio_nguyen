"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useChangeUserRole } from "@/hooks/use-user-mutations";
import { ApiError } from "@/lib/api/types";
import { userRoleFormSchema, type UserRoleFormValues } from "@/lib/validation/user-schema";
import type { UserListItem } from "@/types/user.types";

interface UserRoleDialogProps {
  user: UserListItem;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/**
 * FR-005. Also the D1-sanctioned way to mint a second admin: promoting a
 * user here is what opens BR-002's guard on the current sole admin —
 * demote/deactivate/delete on that account only succeed after this runs.
 */
export function UserRoleDialog({ user, open, onOpenChange }: UserRoleDialogProps) {
  const changeRole = useChangeUserRole();
  const form = useForm<UserRoleFormValues>({
    resolver: zodResolver(userRoleFormSchema),
    defaultValues: { role: user.role },
  });

  useEffect(() => {
    if (open) form.reset({ role: user.role });
  }, [open, user.role, form]);

  function handleSubmit(values: UserRoleFormValues) {
    changeRole.mutate(
      { id: user.id, role: values.role },
      {
        onSuccess: () => {
          toast.success("Role updated");
          onOpenChange(false);
        },
        onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not change this account's role."),
      },
    );
  }

  const busy = changeRole.isPending;

  return (
    <Dialog open={open} onOpenChange={(o) => !busy && onOpenChange(o)}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Change role</DialogTitle>
          <DialogDescription>{user.email} will be granted the selected role&apos;s permissions immediately.</DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4" noValidate>
            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Role</FormLabel>
                  <Select value={field.value} onValueChange={field.onChange}>
                    <FormControl>
                      <SelectTrigger aria-label="Role">
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="admin">Admin</SelectItem>
                      <SelectItem value="user">User</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>
                Cancel
              </Button>
              <Button type="submit" disabled={busy}>
                {busy && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                Save
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
