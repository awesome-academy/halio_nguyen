import type { FieldValues, Path, UseFormSetError } from "react-hook-form";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/types";

/**
 * Field names that always have a real rendered control, either at the top
 * level (the base tour form) or under a `schedules.<i>.` / `images.<i>.`
 * prefix (the row editors). A3's nested-array validators share the same
 * per-row validators as A7/A10, so on create they can return flat,
 * unindexed keys (e.g. "return_date", "image_url", "highlights") that have
 * no matching control at either level — those must fall back to a toast
 * instead of silently vanishing into `setError` on a path nothing renders.
 */
const KNOWN_FIELD_NAMES = new Set([
  "category_id", "title", "description", "itinerary", "destination",
  "duration_days", "duration_nights", "price", "discount_price",
  "max_participants", "thumbnail_url", "inclusions", "exclusions",
  "departure_date", "return_date", "available_slots", "price_override",
  "image_url", "caption", "sort_order",
]);

/**
 * Step 11's shared dotted-path mapper: an `ApiError.fields` entry (e.g.
 * `"departure_date"`) is written onto the form at `${prefix}${field}` (e.g.
 * `"schedules.0.departure_date"` when called from that row's own save
 * handler — R3/FR-402). A prefix-less call maps flat base-field names
 * (title, category_id, ...) directly. Any error with no `fields` map, or
 * whose field name has no matching control anywhere, falls back to a toast
 * so it is never silently swallowed.
 */
export function applyServerFieldErrors<TFieldValues extends FieldValues>(err: unknown, setError: UseFormSetError<TFieldValues>, prefix = ""): void {
  if (err instanceof ApiError && err.fields && Object.keys(err.fields).length > 0) {
    let sawUnknownField = false;
    for (const [field, message] of Object.entries(err.fields)) {
      if (!KNOWN_FIELD_NAMES.has(field)) {
        sawUnknownField = true;
        continue;
      }
      // The field path is assembled at runtime from the server's response, so
      // it cannot be checked against RHF's static Path<TFieldValues> union.
      setError(`${prefix}${field}` as Path<TFieldValues>, { type: "server", message });
    }
    if (sawUnknownField) toast.error(err.message);
    return;
  }
  toast.error(err instanceof ApiError ? err.message : "Something went wrong. Please try again.");
}
