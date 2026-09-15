"use client";

import { useState } from "react";
import { useFieldArray, useFormContext, type Path } from "react-hook-form";
import { Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useCreateTourImage, useDeleteTourImage, useUpdateTourImage } from "@/hooks/use-tour-mutations";
import { tourImageSchema } from "@/lib/validation/tour-schema";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";
import { applyServerFieldErrors } from "./server-error-mapper";
import { TourImageRow } from "./tour-image-row";

interface TourImagesEditorProps {
  mode: "create" | "edit";
  /** Required once mode === "edit" — create-mode has no tour id yet (R2). */
  tourId?: string;
}

/** Create mode: useFieldArray only, submitted with the rest of the form
 * (order = array index, A3). Edit mode: every add/save/remove/reorder calls
 * A7-A9 immediately (step 8, R2's two named branches). */
export function TourImagesEditor({ mode, tourId }: TourImagesEditorProps) {
  const form = useFormContext<TourCreateFormValues>();
  const { fields, append, remove, swap, update } = useFieldArray({ control: form.control, name: "images" });
  const [savingIndex, setSavingIndex] = useState<number | null>(null);

  const createImage = useCreateTourImage(tourId ?? "");
  const updateImage = useUpdateTourImage(tourId ?? "");
  const deleteImage = useDeleteTourImage(tourId ?? "");

  async function handleRemove(index: number) {
    const row = fields[index];
    if (mode === "edit" && row.id) {
      try {
        await deleteImage.mutateAsync(row.id);
      } catch {
        toast.error("Could not remove image.");
        return;
      }
    }
    remove(index);
  }

  async function handleMove(index: number, direction: -1 | 1) {
    const target = index + direction;
    if (target < 0 || target >= fields.length) return;
    const a = fields[index];
    const b = fields[target];
    swap(index, target);
    if (mode !== "edit" || !tourId || !a.id || !b.id) return; // unsaved rows have no server sort_order to persist
    try {
      await Promise.all([
        updateImage.mutateAsync({ imageId: a.id, payload: { caption: a.caption || undefined, sort_order: target } }),
        updateImage.mutateAsync({ imageId: b.id, payload: { caption: b.caption || undefined, sort_order: index } }),
      ]);
    } catch {
      swap(index, target); // R5: roll back the optimistic reorder on failure
      toast.error("Could not reorder images.");
    }
  }

  async function handleSave(index: number) {
    const row = form.getValues(`images.${index}`);
    const parsed = tourImageSchema.safeParse(row);
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        // The path is assembled at runtime from zod's issue.path, so it
        // cannot be checked against RHF's static Path<T> union.
        form.setError(`images.${index}.${issue.path.join(".")}` as Path<TourCreateFormValues>, { type: "validation", message: issue.message });
      }
      return;
    }
    setSavingIndex(index);
    try {
      if (row.id) {
        const updated = await updateImage.mutateAsync({ imageId: row.id, payload: { caption: parsed.data.caption, sort_order: row.sort_order ?? index } });
        update(index, { id: updated.id, image_url: updated.image_url, caption: updated.caption ?? "", sort_order: updated.sort_order });
      } else {
        const created = await createImage.mutateAsync({ image_url: parsed.data.image_url, caption: parsed.data.caption });
        update(index, { id: created.id, image_url: created.image_url, caption: created.caption ?? "", sort_order: created.sort_order });
      }
      toast.success("Image saved");
    } catch (err) {
      applyServerFieldErrors(err, form.setError, `images.${index}.`);
    } finally {
      setSavingIndex(null);
    }
  }

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <h2 className="font-semibold">Gallery</h2>
      {fields.length === 0 && <p className="text-sm text-muted-foreground">No images yet.</p>}
      {fields.map((field, index) => (
        <TourImageRow
          key={field.id}
          index={index}
          mode={mode}
          isSaved={mode === "edit" && Boolean(field.id)}
          isFirst={index === 0}
          isLast={index === fields.length - 1}
          isSaving={savingIndex === index}
          onRemove={() => handleRemove(index)}
          onMove={(direction) => handleMove(index, direction)}
          onSave={mode === "edit" ? () => handleSave(index) : undefined}
        />
      ))}
      <Button type="button" variant="outline" size="sm" onClick={() => append({ image_url: "", caption: "" })}>
        <Plus className="h-4 w-4 mr-2" /> Add image
      </Button>
    </div>
  );
}
