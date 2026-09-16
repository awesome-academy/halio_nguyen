import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery, Paginated } from "./types";
import type { UserDetail, UserListItem, UserUpdatePayload } from "@/types/user.types";

const BASE = "/api/v1/admin/users";

export interface UserListQuery extends ListQuery {
  role?: string;
  is_active?: boolean;
}

// A1
export function listUsers(query: UserListQuery): Promise<Paginated<UserListItem>> {
  return apiFetch<Paginated<UserListItem>>(`${BASE}${toQueryString(query)}`);
}

// A2 — booking_count/review_count are lifetime totals, not BR-003's active-booking gate.
export function getUser(id: string): Promise<UserDetail> {
  return apiFetch<UserDetail>(`${BASE}/${id}`);
}

// A3 — at least one field; the guard pipeline (BR-001 self-lockout, BR-002
// last-admin) runs server-side and answers 403/409 with the message to show.
export function updateUser(id: string, payload: UserUpdatePayload): Promise<UserListItem> {
  return apiFetch<UserListItem>(`${BASE}/${id}`, { method: "PATCH", body: payload });
}

// A4 — soft-delete; blocked 409 by BR-003 when active bookings remain.
export function deleteUser(id: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${id}`, { method: "DELETE" });
}
