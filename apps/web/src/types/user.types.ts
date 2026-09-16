export type UserRole = 'admin' | 'user';

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  avatar_url?: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserBankAccount {
  id: string;
  user_id: string;
  bank_name: string;
  bank_code?: string;
  account_number: string;
  account_holder_name: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

// --- Admin portal DTOs (F005) --------------------------------------------
// Mirrors apps/api/internal/domain's user list/detail shapes (A1/A2).
// `password_hash` is `json:"-"` on the Go struct and MUST never be declared
// here — see phase-07-platform-user-management.md's Security Considerations.

/** One row of GET /api/v1/admin/users (FR-001/FR-002). */
export interface UserListItem {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  avatar_url?: string;
  role: UserRole;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

/** GET /api/v1/admin/users/:id — full profile plus the two lifetime history
 * counts (FR-003). These are lifetime totals, distinct from BR-003's
 * active-booking count used only to gate the delete. */
export interface UserDetail extends UserListItem {
  booking_count: number;
  review_count: number;
}

/** PATCH /api/v1/admin/users/:id body — at least one field is required
 * (FR-004/FR-005); the server rejects an empty patch with 422. */
export interface UserUpdatePayload {
  is_active?: boolean;
  role?: UserRole;
}
