import { apiFetch } from "./client";

/** The admin profile every session-aware screen reads (never password_hash). */
export interface AdminProfile {
  id: string;
  email: string;
  full_name: string;
  role: string;
}

interface LoginResponse {
  user: AdminProfile;
  redirect_to: string;
}

/**
 * Posts credentials to the login endpoint. from is the same-origin admin
 * path middleware.ts recorded before redirecting here (DEC-001) — the
 * server re-validates it and only ever returns a same-origin redirect_to.
 */
export async function login(email: string, password: string, from?: string): Promise<LoginResponse> {
  return apiFetch<LoginResponse>("/api/v1/admin/auth/login", {
    method: "POST",
    body: { email, password, from },
  });
}

export async function logout(): Promise<void> {
  await apiFetch<void>("/api/v1/admin/auth/logout", { method: "POST" });
}

export async function getSession(): Promise<AdminProfile> {
  return apiFetch<AdminProfile>("/api/v1/admin/auth/me");
}
