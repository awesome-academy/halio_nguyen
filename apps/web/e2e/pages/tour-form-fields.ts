/**
 * Accessible names for the tour create/edit form, read directly from
 * `apps/web/src/app/(admin)/admin/tours/_components/tour-form-*.tsx` at
 * implementation time (decisions.md D4 forbids guessing them). Kept in its
 * own file so `tour-form-page.ts` stays under the 200-line cap once the
 * three form sections plus the image/schedule editors are all wired in.
 */

/** `tour-form-basic-section.tsx` + `tour-form-pricing-section.tsx` + `tour-form-content-section.tsx`. */
export const TOUR_FIELDS = {
  title: "Title",
  category: "Category",
  destination: "Destination",
  description: "Description",
  itinerary: "Itinerary",
  price: "Price (VND)",
  discountPrice: "Discounted price (VND)",
  durationDays: "Duration (days)",
  durationNights: "Duration (nights)",
  maxParticipants: "Max participants",
  thumbnailUrl: "Thumbnail URL",
  inclusions: "Inclusions",
  exclusions: "Exclusions",
} as const;

/** `tour-images-editor.tsx` / `tour-image-row.tsx`. Save/Move/Remove are edit-mode only. */
export const TOUR_IMAGE_FIELDS = {
  addImage: "Add image",
  saveImage: (n: number) => `Save image ${n}`,
  removeImage: (n: number) => `Remove image ${n}`,
  moveImageUp: (n: number) => `Move image ${n} up`,
  moveImageDown: (n: number) => `Move image ${n} down`,
} as const;

/** `tour-schedules-editor.tsx` / `tour-schedule-row.tsx`. Date/slot/price fields use a plain `aria-label`. */
export const TOUR_SCHEDULE_FIELDS = {
  addSchedule: "Add schedule",
  saveSchedule: (n: number) => `Save schedule ${n}`,
  removeSchedule: (n: number) => `Remove schedule ${n}`,
  departureDate: "Departure date",
  returnDate: "Return date",
  availableSlots: "Available slots",
  priceOverride: "Price override",
} as const;
