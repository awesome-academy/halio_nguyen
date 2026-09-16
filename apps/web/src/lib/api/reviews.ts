import { apiFetch } from "./client";
import { toQueryString } from "./query-string";
import type { ListQuery, Paginated } from "./types";
import type { CommentNode, ModerationStatus, Review, ReviewCategory, ReviewDetail, ReviewListItem } from "@/types/review.types";

const BASE = "/api/v1/admin/reviews";
const CATEGORY_BASE = "/api/v1/admin/review-categories";

export interface ReviewListQuery extends ListQuery {
  status?: string;
  review_category_id?: string;
}

// A1
export function listReviews(query: ReviewListQuery): Promise<Paginated<ReviewListItem>> {
  return apiFetch<Paginated<ReviewListItem>>(`${BASE}${toQueryString(query)}`);
}

// A2 — 404 when the review is missing or soft-deleted.
export function getReview(id: string): Promise<ReviewDetail> {
  return apiFetch<ReviewDetail>(`${BASE}/${id}`);
}

// A3 — SM-001 admits only published/hidden; "draft" is rejected 422 server-side.
export function updateReviewStatus(id: string, status: ModerationStatus): Promise<Review> {
  return apiFetch<Review>(`${BASE}/${id}/status`, { method: "PATCH", body: { status } });
}

// A4 — soft-delete; terminal, no restore path (§5 F006 default).
export function deleteReview(id: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${id}`, { method: "DELETE" });
}

// A5 — scoped by review_id + comment_id; a mismatched pair 404s server-side.
export function updateCommentVisibility(reviewId: string, commentId: string, isHidden: boolean): Promise<CommentNode> {
  return apiFetch<CommentNode>(`${BASE}/${reviewId}/comments/${commentId}/visibility`, {
    method: "PATCH",
    body: { is_hidden: isHidden },
  });
}

// A6 — soft-delete; replies survive independently (BR-003).
export function deleteComment(reviewId: string, commentId: string): Promise<void> {
  return apiFetch<void>(`${BASE}/${reviewId}/comments/${commentId}`, { method: "DELETE" });
}

// A7 — bare array (not paginated); the seeded review-category lookup that
// feeds the list filter. Never Phase 3's tour-category endpoint.
export function listReviewCategories(): Promise<ReviewCategory[]> {
  return apiFetch<ReviewCategory[]>(CATEGORY_BASE);
}
