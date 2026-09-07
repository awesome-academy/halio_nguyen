import { User } from './user.types';

export interface ReviewCategory {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

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
  status: 'draft' | 'published' | 'hidden';
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
