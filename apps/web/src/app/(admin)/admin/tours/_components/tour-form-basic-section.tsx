"use client";

import { useFormContext } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { slugify } from "@/lib/validation/tour-schema";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";
import type { CategoryListItem } from "@/types/category.types";

interface TourFormBasicSectionProps {
  categories: CategoryListItem[];
  categoriesLoading: boolean;
}

/** Title, destination, description, itinerary, category picker (R4), and a
 * read-only slug preview — the slug itself is always server-derived (FR-002)
 * and is never part of the submitted payload. */
export function TourFormBasicSection({ categories, categoriesLoading }: TourFormBasicSectionProps) {
  const form = useFormContext<TourCreateFormValues>();
  const title = form.watch("title");

  return (
    <div className="rounded-lg border bg-card p-4 space-y-4">
      <h2 className="font-semibold">Basic information</h2>

      <FormField
        control={form.control}
        name="title"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Title</FormLabel>
            <FormControl>
              <Input autoFocus {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <div className="text-xs text-muted-foreground">
        URL slug (auto-generated): <span className="font-mono">/{slugify(title || "")}</span>
      </div>

      <FormField
        control={form.control}
        name="category_id"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Category</FormLabel>
            <Select value={field.value} onValueChange={field.onChange} disabled={categoriesLoading}>
              <FormControl>
                <SelectTrigger>
                  <SelectValue placeholder={categoriesLoading ? "Loading…" : "Select a category"} />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {categories.map((c) => (
                  <SelectItem key={c.id} value={c.id}>
                    {c.name}
                    {!c.is_active ? " (inactive)" : ""}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={form.control}
        name="destination"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Destination</FormLabel>
            <FormControl>
              <Input {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={form.control}
        name="description"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Description</FormLabel>
            <FormControl>
              <Textarea rows={4} {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={form.control}
        name="itinerary"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Itinerary</FormLabel>
            <FormControl>
              <Textarea rows={4} {...field} value={field.value ?? ""} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}
