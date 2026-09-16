"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { refreshRevenue } from "@/lib/api/revenue";
import { revenueKeys } from "@/lib/api/query-keys";
import { ApiError } from "@/lib/api/types";

/**
 * A3. The API answers 202 before the refresh finishes and there is no push
 * channel back, so this invalidates both reports and leaves it at that:
 * they re-fetch now and again whenever the admin returns. Polling is
 * deliberately not added — the spec scopes this to user-triggered re-fetch.
 */
export function RefreshRevenueButton() {
  const queryClient = useQueryClient();

  const trigger = useMutation({
    mutationFn: refreshRevenue,
    onSuccess: () => {
      toast.success("Refresh started", {
        description: "The reports rebuild in the background. Re-check in a moment for the new figures.",
      });
      void queryClient.invalidateQueries({ queryKey: revenueKeys.all });
    },
    onError: (err) => {
      // 409 is the expected answer while another refresh is running — it is
      // rejected, not queued, so say so rather than showing a failure.
      if (err instanceof ApiError && err.status === 409) {
        toast.info("A refresh is already running", { description: "Wait for it to finish, then try again." });
        return;
      }
      toast.error(err instanceof ApiError ? err.message : "Could not start the refresh. Please try again.");
    },
  });

  return (
    <Button size="sm" onClick={() => trigger.mutate()} disabled={trigger.isPending}>
      <RefreshCw className={`mr-2 h-3.5 w-3.5 ${trigger.isPending ? "animate-spin" : ""}`} />
      Trigger Batch Refresh
    </Button>
  );
}
