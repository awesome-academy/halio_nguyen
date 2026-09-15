"use client";

import { useFieldArray, useFormContext } from "react-hook-form";
import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";

/** highlights[] (useFieldArray), inclusions, exclusions. */
export function TourFormContentSection() {
  const form = useFormContext<TourCreateFormValues>();
  const { fields, append, remove } = useFieldArray({ control: form.control, name: "highlights" });

  return (
    <div className="rounded-lg border bg-card p-4 space-y-4">
      <h2 className="font-semibold">Highlights &amp; policies</h2>

      <div className="space-y-2">
        <Label>Highlights</Label>
        {fields.map((field, index) => (
          <FormField
            key={field.id}
            control={form.control}
            name={`highlights.${index}.value`}
            render={({ field: inputField }) => (
              <FormItem>
                <div className="flex items-center gap-2">
                  <FormControl>
                    <Input {...inputField} />
                  </FormControl>
                  <Button type="button" variant="ghost" size="icon" className="h-8 w-8 shrink-0" aria-label={`Remove highlight ${index + 1}`} onClick={() => remove(index)}>
                    <X className="h-4 w-4" />
                  </Button>
                </div>
                <FormMessage />
              </FormItem>
            )}
          />
        ))}
        <Button type="button" variant="outline" size="sm" disabled={fields.length >= 20} onClick={() => append({ value: "" })}>
          <Plus className="h-4 w-4 mr-2" /> Add highlight
        </Button>
      </div>

      <FormField
        control={form.control}
        name="inclusions"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Inclusions</FormLabel>
            <FormControl>
              <Textarea rows={3} {...field} value={field.value ?? ""} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={form.control}
        name="exclusions"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Exclusions</FormLabel>
            <FormControl>
              <Textarea rows={3} {...field} value={field.value ?? ""} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}
