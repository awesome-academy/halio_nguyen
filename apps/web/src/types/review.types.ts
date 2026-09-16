import { User } from './user.types';

export interface ReviewCategory {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

/** SM-001 admits only published <-> hidden as moderation targets; `draft`
 * is display-only (it arrives from the authoring flow, out of scope). */
export type ReviewStatus = 'draft' | 'published' | 'hidden';
export type ModerationStatus = Extract<ReviewStatus, 'published' | 'hidden'>;

export interface Review {
  id: string;
  user_id: string;
  user?: User;
  review_category_id: string;
  review_category?: ReviewCategory;
  title: string;
  slug: string;
  content: string;
  thumbnail_url?: string;
  like_count: number;
  comment_count: number;
  status: ReviewStatus;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  review_id: string;
  user_id: string;
  user?: User;
  parent_id?: string;
  replies?: Comment[];
  content: string;
  is_hidden: boolean;
  created_at: string;
  updated_at: string;
}

export interface TourRating {
  id: string;
  tour_id: string;
  user_id: string;
  user?: User;
  booking_id: string;
  score: number;
  comment?: string;
  created_at: string;
}

/** A1 row — no `content`: the list renders titles only. */
export interface ReviewListItem {
  id: string;
  title: string;
  slug: string;
  status: ReviewStatus;
  like_count: number;
  comment_count: number;
  created_at: string;
  updated_at: string;
  user_id: string;
  author_name: string;
  author_avatar_url?: string;
  review_category_id: string;
  category_name: string;
}

export interface CommentAuthor {
  id: string;
  full_name: string;
  avatar_url?: string;
}

/** A soft-deleted node is REDACTED server-side: `content` and `user` are
 * simply absent. `is_deleted` is the discriminant a component must check
 * before it may read `content`/`user` off a node. */
export interface CommentNode {
  id: string;
  parent_id?: string;
  content?: string; // absent when is_deleted
  is_hidden: boolean;
  is_deleted: boolean;
  created_at: string;
  user?: CommentAuthor; // absent when is_deleted
  replies: CommentNode[];
}

export interface ReviewDetail extends Review {
  author_name: string;
  author_avatar_url?: string;
  category_name: string;
  comments: CommentNode[];
}
