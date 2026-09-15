"use client";

import { Archive, CheckCircle2, Loader2, RotateCcw } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useUpdateTourStatus } from "@/hooks/use-tour-mutations";
import { ApiError } from "@/lib/api/types";
import type { TourStatus } from "@/types/tour.types";
import { allowedTourActions } from "./tour-status-rules";

interface TourStatusActionsProps {
  tourId: string;
  status: TourStatus;
}

const STATUS_LABEL: Record<TourStatus, string> = { draft: "Draft", published: "Published", archived: "Archived" };

/** DEC-001 on the form header — A5 is the single writer of tours.status. */
export function TourStatusActions({ tourId, status }: TourStatusActionsProps) {
  const updateStatus = useUpdateTourStatus();
  const actions = allowedTourActions(status);

  function change(next: TourStatus) {
    updateStatus.mutate(
      { id: tourId, status: next },
      { onError: (err) => toast.error(err instanceof ApiError ? err.message : "Could not change status.") },
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Badge variant={status === "published" ? "default" : status === "archived" ? "secondary" : "outline"}>{STATUS_LABEL[status]}</Badge>
      {updateStatus.isPending && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
      {actions.includes("publish") && (
        <Button type="button" size="sm" variant="outline" disabled={updateStatus.isPending} onClick={() => change("published")}>
          <CheckCircle2 className="h-4 w-4 mr-2" /> Publish
        </Button>
      )}
      {actions.includes("archive") && (
        <Button type="button" size="sm" variant="outline" disabled={updateStatus.isPending} onClick={() => change("archived")}>
          <Archive className="h-4 w-4 mr-2" /> Archive
        </Button>
      )}
      {actions.includes("reactivate") && (
        <Button type="button" size="sm" variant="outline" disabled={updateStatus.isPending} onClick={() => change("published")}>
          <RotateCcw className="h-4 w-4 mr-2" /> Reactivate
        </Button>
      )}
    </div>
  );
}
