import { z } from "zod";

// Mirrors apps/api/internal/service/tour_service_write.go (buildTour),
// tour_image_service.go (buildTourImage) and tour_schedule_service.go
// (buildTourSchedule). Client-side convenience only — the server
// re-validates everything (Security Considerations, phase-05).

/** http(s) absolute URLs only — implicitly rejects javascript:/data: schemes,
 * which never match this shape (Security: rendered into <img src>). */
const HTTP_URL_SHAPE = /^https?:\/\/\S+$/i;

function httpUrl(message: string) {
  return z
    .string()
    .trim()
    .refine((v) => HTTP_URL_SHAPE.test(v), message);
}

const optionalNumber = (message: string) =>
  z.coerce
    .number()
    .min(0, message)
    .optional()
    .or(z.literal("").transform(() => undefined));

export const tourImageSchema = z.object({
  id: z.string().uuid().optional(),
  image_url: httpUrl("Please enter a valid http(s) URL."),
  caption: z.string().trim().optional(),
  sort_order: z.number().int().min(0).optional(),
});
export type TourImageFormValues = z.infer<typeof tourImageSchema>;

export const tourScheduleSchema = z
  .object({
    id: z.string().uuid().optional(),
    departure_date: z.string().min(1, "Departure date is required."),
    return_date: z.string().min(1, "Return date is required."),
    available_slots: z.coerce.number().int().min(0, "Available slots must be zero or more."),
    price_override: optionalNumber("Price override must be zero or more."),
    status: z.enum(["open", "closed", "cancelled"]).optional(),
  })
  .superRefine((val, ctx) => {
    if (val.departure_date && val.return_date && val.return_date < val.departure_date) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["return_date"], message: "Return date must be on or after the departure date." });
    }
  });
export type TourScheduleFormValues = z.infer<typeof tourScheduleSchema>;

const tourBaseShape = {
  category_id: z.string().uuid("Please select a category."),
  title: z.string().trim().min(1, "Title is required.").max(255, "Title must be 255 characters or fewer."),
  destination: z.string().trim().min(1, "Destination is required.").max(255, "Destination must be 255 characters or fewer."),
  description: z.string().trim().min(1, "Description is required."),
  itinerary: z.string().trim().optional(),
  duration_days: z.coerce.number().int().min(1, "Duration (days) must be greater than zero."),
  duration_nights: z.coerce.number().int().min(0, "Duration (nights) must be zero or more."),
  price: z.coerce.number().min(0, "Price must be zero or more."),
  discount_price: optionalNumber("Discounted price must be zero or more."),
  max_participants: z.coerce.number().int().min(1, "Max participants must be greater than zero."),
  thumbnail_url: httpUrl("Please enter a valid http(s) URL.").optional().or(z.literal("")),
  highlights: z
    .array(z.object({ value: z.string().trim().max(200, "Each highlight must be 200 characters or fewer.") }))
    .max(20, "No more than 20 highlights are allowed."),
  inclusions: z.string().trim().optional(),
  exclusions: z.string().trim().optional(),
};

const tourBaseObject = z.object(tourBaseShape);

/** BR-001: discount_price, when set, must not exceed price. Shared between
 * tourBaseSchema (edit) and tourCreateSchema so the rule is defined once. */
function checkDiscountPrice(val: { price: number; discount_price?: number }, ctx: z.RefinementCtx) {
  if (val.discount_price !== undefined && val.discount_price > val.price) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["discount_price"], message: "Discounted price cannot exceed the regular price." });
  }
}

/** Edit-mode form (PUT /tours/:id — A4 excludes images/schedules/status). */
export const tourBaseSchema = tourBaseObject.superRefine(checkDiscountPrice);
export type TourBaseFormValues = z.infer<typeof tourBaseObject>;

/** Create-mode only: base fields plus the two nested arrays submitted in one
 * POST (R2). Duplicate departure dates are rejected client-side too, ahead
 * of the server's 409 (FR-402). */
export const tourCreateSchema = tourBaseObject
  .extend({
    images: z.array(tourImageSchema),
    schedules: z.array(tourScheduleSchema),
  })
  .superRefine((val, ctx) => {
    checkDiscountPrice(val, ctx);
    const seen = new Map<string, number>();
    val.schedules.forEach((s, i) => {
      if (!s.departure_date) return;
      if (seen.has(s.departure_date)) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, path: ["schedules", i, "departure_date"], message: "Duplicate departure date." });
      } else {
        seen.set(s.departure_date, i);
      }
    });
  });
export type TourCreateFormValues = z.infer<typeof tourCreateSchema>;

/** Client-side twin of service.Slugify — preview only; slug is always
 * server-derived and never part of the submitted payload. */
export function slugify(input: string): string {
  return input
    .replace(/đ/g, "d")
    .replace(/Đ/g, "D")
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 280)
    .replace(/-+$/g, "");
}
