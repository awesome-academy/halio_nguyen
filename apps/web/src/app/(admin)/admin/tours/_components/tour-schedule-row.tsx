"use client";

import { useFormContext } from "react-hook-form";
import { Ban, Loader2, PlayCircle, Save, StopCircle, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";
import type { ScheduleStatus } from "@/types/tour.types";
import { allowedScheduleActions } from "./tour-status-rules";

interface TourScheduleRowProps {
  index: number;
  mode: "create" | "edit";
  status?: ScheduleStatus;
  isSaving?: boolean;
  isChangingStatus?: boolean;
  onRemove: () => void;
  onSave?: () => void;
  onStatusChange?: (status: "open" | "closed" | "cancelled") => void;
}

const STATUS_LABEL: Record<ScheduleStatus, string> = { open: "Open", closed: "Closed", cancelled: "Cancelled" };

/** DEC-002's row-level actions — Close/Reopen/Cancel come straight from allowedScheduleActions. */
export function TourScheduleRow({ index, mode, status, isSaving, isChangingStatus, onRemove, onSave, onStatusChange }: TourScheduleRowProps) {
  const form = useFormContext<TourCreateFormValues>();
  const actions = status ? allowedScheduleActions(status) : ["edit" as const];

  return (
    <div className="rounded-md border p-3 space-y-2">
      <div className="flex items-center justify-between">
        {status && <Badge variant={status === "cancelled" ? "secondary" : "outline"}>{STATUS_LABEL[status]}</Badge>}
        <div className="flex items-center gap-1 ml-auto">
          {mode === "edit" && actions.includes("close") && onStatusChange && (
            <Button type="button" variant="ghost" size="sm" disabled={isChangingStatus} onClick={() => onStatusChange("closed")}>
              <StopCircle className="h-3.5 w-3.5 mr-1" /> Close
            </Button>
          )}
          {mode === "edit" && actions.includes("reopen") && onStatusChange && (
            <Button type="button" variant="ghost" size="sm" disabled={isChangingStatus} onClick={() => onStatusChange("open")}>
              <PlayCircle className="h-3.5 w-3.5 mr-1" /> Reopen
            </Button>
          )}
          {mode === "edit" && actions.includes("cancel") && onStatusChange && (
            <Button type="button" variant="ghost" size="sm" className="text-destructive" disabled={isChangingStatus} onClick={() => onStatusChange("cancelled")}>
              <Ban className="h-3.5 w-3.5 mr-1" /> Cancel
            </Button>
          )}
          {mode === "edit" && onSave && (
            <Button type="button" variant="outline" size="icon" className="h-8 w-8" aria-label={`Save schedule ${index + 1}`} disabled={isSaving} onClick={onSave}>
              {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            </Button>
          )}
          <Button type="button" variant="ghost" size="icon" className="h-8 w-8 text-destructive" aria-label={`Remove schedule ${index + 1}`} onClick={onRemove}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <FormField
          control={form.control}
          name={`schedules.${index}.departure_date`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input type="date" aria-label="Departure date" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`schedules.${index}.return_date`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input type="date" aria-label="Return date" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`schedules.${index}.available_slots`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input type="number" min={0} placeholder="Available slots" aria-label="Available slots" {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`schedules.${index}.price_override`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input type="number" min={0} placeholder="Price override (optional)" aria-label="Price override" {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>
    </div>
  );
}
