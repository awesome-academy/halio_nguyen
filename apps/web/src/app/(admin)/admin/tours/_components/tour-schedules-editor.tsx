"use client";

import { useState } from "react";
import { useFieldArray, useFormContext, type Path } from "react-hook-form";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useCreateTourSchedule, useDeleteTourSchedule, useUpdateTourSchedule, useUpdateTourScheduleStatus } from "@/hooks/use-tour-mutations";
import { ApiError } from "@/lib/api/types";
import { tourScheduleSchema } from "@/lib/validation/tour-schema";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";
import { applyServerFieldErrors } from "./server-error-mapper";
import { TourScheduleRow } from "./tour-schedule-row";

interface TourSchedulesEditorProps {
  mode: "create" | "edit";
  /** Required once mode === "edit" — create-mode has no tour id yet (R2). */
  tourId?: string;
}

/** Create mode: useFieldArray only, submitted with the rest of the form
 * (A3). Edit mode: add/save/remove call A10/A11/A13 immediately and status
 * changes go through A12 (step 9, R2's two named branches). */
export function TourSchedulesEditor({ mode, tourId }: TourSchedulesEditorProps) {
  const form = useFormContext<TourCreateFormValues>();
  const { fields, append, remove, update } = useFieldArray({ control: form.control, name: "schedules" });
  const [savingIndex, setSavingIndex] = useState<number | null>(null);
  const [statusIndex, setStatusIndex] = useState<number | null>(null);

  const createSchedule = useCreateTourSchedule(tourId ?? "");
  const updateSchedule = useUpdateTourSchedule(tourId ?? "");
  const updateStatus = useUpdateTourScheduleStatus(tourId ?? "");
  const deleteSchedule = useDeleteTourSchedule(tourId ?? "");

  async function handleRemove(index: number) {
    const row = fields[index];
    if (mode === "edit" && row.id) {
      try {
        await deleteSchedule.mutateAsync(row.id);
      } catch (err) {
        // Mirrors tour-delete-dialog's D4 handling: surface the server's own
        // message (e.g. the active-booking count) instead of a generic string.
        toast.error(err instanceof ApiError ? err.message : "Could not remove schedule.");
        return;
      }
    }
    remove(index);
  }

  async function handleSave(index: number) {
    const row = form.getValues(`schedules.${index}`);
    const parsed = tourScheduleSchema.safeParse(row);
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        // The path is assembled at runtime from zod's issue.path, so it
        // cannot be checked against RHF's static Path<T> union.
        form.setError(`schedules.${index}.${issue.path.join(".")}` as Path<TourCreateFormValues>, { type: "validation", message: issue.message });
      }
      return;
    }
    const payload = {
      departure_date: parsed.data.departure_date,
      return_date: parsed.data.return_date,
      available_slots: parsed.data.available_slots,
      price_override: parsed.data.price_override,
    };
    setSavingIndex(index);
    try {
      if (row.id) {
        const updated = await updateSchedule.mutateAsync({ scheduleId: row.id, payload });
        update(index, { ...updated });
      } else {
        const created = await createSchedule.mutateAsync(payload);
        update(index, { ...created });
      }
      toast.success("Schedule saved");
    } catch (err) {
      // FR-402: a duplicate departure_date 409 lands on this exact row's date field.
      applyServerFieldErrors(err, form.setError, `schedules.${index}.`);
    } finally {
      setSavingIndex(null);
    }
  }

  async function handleStatusChange(index: number, status: "open" | "closed" | "cancelled") {
    const row = fields[index];
    if (!row.id) return;
    setStatusIndex(index);
    try {
      const updated = await updateStatus.mutateAsync({ scheduleId: row.id, status });
      update(index, { ...row, status: updated.status });
    } catch (err) {
      applyServerFieldErrors(err, form.setError, `schedules.${index}.`);
    } finally {
      setStatusIndex(null);
    }
  }

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <h2 className="font-semibold">Departure schedules</h2>
      {fields.length === 0 && <p className="text-sm text-muted-foreground">No schedules yet.</p>}
      {fields.map((field, index) => (
        <TourScheduleRow
          key={field.id}
          index={index}
          mode={mode}
          status={field.status}
          isSaving={savingIndex === index}
          isChangingStatus={statusIndex === index}
          onRemove={() => handleRemove(index)}
          onSave={mode === "edit" ? () => handleSave(index) : undefined}
          onStatusChange={mode === "edit" && field.id ? (status) => handleStatusChange(index, status) : undefined}
        />
      ))}
      <Button type="button" variant="outline" size="sm" onClick={() => append({ departure_date: "", return_date: "", available_slots: 0, status: "open" })}>
        <Plus className="h-4 w-4 mr-2" /> Add schedule
      </Button>
    </div>
  );
}
