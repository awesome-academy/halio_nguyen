import { z } from "zod";

/** Mirrors domain.RoleAdmin/domain.RoleUser in apps/api/internal/domain/user.go. */
export const userRoleSchema = z.enum(["admin", "user"]);

/** FR-005's role dialog. Server-side validation (422 on anything else) is
 * the real gate; this only keeps a bad value out of the request body. */
export const userRoleFormSchema = z.object({
  role: userRoleSchema,
});

export type UserRoleFormValues = z.infer<typeof userRoleFormSchema>;
