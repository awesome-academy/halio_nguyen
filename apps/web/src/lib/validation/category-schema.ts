import { z } from "zod";

const SLUG_SHAPE = /^[a-z0-9]+(-[a-z0-9]+)*$/;

/** Mirrors the server's rules in apps/api/internal/service/category_rules.go. */
export const categorySchema = z.object({
  name: z.string().trim().min(1, "Name is required.").max(100, "Name must be 100 characters or fewer."),
  slug: z
    .string()
    .trim()
    .min(1, "Slug is required.")
    .max(120, "Slug must be 120 characters or fewer.")
    .regex(SLUG_SHAPE, "Slug must be lowercase letters, numbers, and hyphens only."),
  description: z.string().trim().optional(),
  image_url: z
    .string()
    .trim()
    .optional()
    .refine((v) => !v || /^https?:\/\/\S+$/i.test(v), "Please enter a valid http(s) URL."),
  sort_order: z.coerce.number().int().min(0).optional(),
  is_active: z.boolean(),
});

export type CategoryFormValues = z.infer<typeof categorySchema>;

/** Client-side twin of service.Slugify — good enough for live preview; the server derives the real one when slug is blank. */
export function slugify(input: string): string {
  return input
    .replace(/đ/g, "d")
    .replace(/Đ/g, "D")
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 120)
    .replace(/-+$/g, "");
}
