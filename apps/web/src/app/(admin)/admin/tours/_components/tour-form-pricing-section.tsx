"use client";

import { useFormContext } from "react-hook-form";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import type { TourCreateFormValues } from "@/lib/validation/tour-schema";

/** price, discount_price, duration_days/nights, max_participants, thumbnail_url. */
export function TourFormPricingSection() {
  const form = useFormContext<TourCreateFormValues>();

  return (
    <div className="rounded-lg border bg-card p-4 space-y-4">
      <h2 className="font-semibold">Pricing &amp; capacity</h2>

      <div className="grid grid-cols-2 gap-4">
        <FormField
          control={form.control}
          name="price"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Price (VND)</FormLabel>
              <FormControl>
                <Input type="number" min={0} step="1000" {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="discount_price"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Discounted price (VND)</FormLabel>
              <FormControl>
                <Input type="number" min={0} step="1000" placeholder="none" {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <FormField
          control={form.control}
          name="duration_days"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Duration (days)</FormLabel>
              <FormControl>
                <Input type="number" min={1} {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="duration_nights"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Duration (nights)</FormLabel>
              <FormControl>
                <Input type="number" min={0} {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="max_participants"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Max participants</FormLabel>
              <FormControl>
                <Input type="number" min={1} {...field} value={field.value ?? ""} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <FormField
        control={form.control}
        name="thumbnail_url"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Thumbnail URL</FormLabel>
            <FormControl>
              <Input type="url" placeholder="https://…" {...field} value={field.value ?? ""} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
}
