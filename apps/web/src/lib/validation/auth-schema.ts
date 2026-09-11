import { z } from "zod";

/**
 * Shared by the login form (client-side required checks, FR-201) and the
 * request payload sent to POST /api/v1/admin/auth/login.
 */
export const loginSchema = z.object({
  email: z.string().min(1, "Email is required"),
  password: z.string().min(1, "Password is required"),
});

export type LoginFormValues = z.infer<typeof loginSchema>;
