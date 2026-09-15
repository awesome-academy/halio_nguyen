"use client";

import { useFormContext } from "react-hook-form";
import { ArrowDown, ArrowUp, Loader2, Save, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";

interface TourImageRowProps {
  index: number;
  mode: "create" | "edit";
  /** Whether this row already exists on the server (A8 cannot change image_url). */
  isSaved: boolean;
  isFirst: boolean;
  isLast: boolean;
  isSaving?: boolean;
  onRemove: () => void;
  onMove: (direction: -1 | 1) => void;
  /** Edit-mode only — create-mode rows submit as part of the whole form. */
  onSave?: () => void;
}

export function TourImageRow({ index, mode, isSaved, isFirst, isLast, isSaving, onRemove, onMove, onSave }: TourImageRowProps) {
  const form = useFormContext<TourCreateFormValues>();

  return (
    <div className="flex items-start gap-2 rounded-md border p-3">
      <div className="flex flex-col shrink-0">
        <Button type="button" variant="ghost" size="icon" className="h-6 w-6" aria-label={`Move image ${index + 1} up`} disabled={isFirst} onClick={() => onMove(-1)}>
          <ArrowUp className="h-3.5 w-3.5" />
        </Button>
        <Button type="button" variant="ghost" size="icon" className="h-6 w-6" aria-label={`Move image ${index + 1} down`} disabled={isLast} onClick={() => onMove(1)}>
          <ArrowDown className="h-3.5 w-3.5" />
        </Button>
      </div>

      <div className="flex-1 space-y-2">
        <FormField
          control={form.control}
          name={`images.${index}.image_url`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="https://…" disabled={mode === "edit" && isSaved} {...field} />
              </FormControl>
              {mode === "edit" && isSaved && <p className="text-xs text-muted-foreground">The image URL cannot be changed after upload — remove and re-add instead.</p>}
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name={`images.${index}.caption`}
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <Input placeholder="Caption (optional)" {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className="flex flex-col gap-1 shrink-0">
        {mode === "edit" && onSave && (
          <Button type="button" variant="outline" size="icon" className="h-8 w-8" aria-label={`Save image ${index + 1}`} disabled={isSaving} onClick={onSave}>
            {isSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          </Button>
        )}
        <Button type="button" variant="ghost" size="icon" className="h-8 w-8 text-destructive" aria-label={`Remove image ${index + 1}`} onClick={onRemove}>
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
