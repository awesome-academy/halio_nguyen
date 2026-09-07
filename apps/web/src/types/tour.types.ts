export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  image_url?: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TourImage {
  id: string;
  tour_id: string;
  image_url: string;
  caption?: string;
  sort_order: number;
}

export interface TourSchedule {
  id: string;
  tour_id: string;
  departure_date: string;
  return_date: string;
  available_slots: number;
  price_override?: number;
  status: 'open' | 'closed' | 'cancelled';
}

export interface Tour {
  id: string;
  category_id: string;
  category?: Category;
  title: string;
  slug: string;
  description: string;
  itinerary?: string;
  destination: string;
  duration_days: number;
  duration_nights: number;
  price: number;
  discount_price?: number;
  max_participants: number;
  thumbnail_url?: string;
  highlights?: string[];
  inclusions?: string;
  exclusions?: string;
  status: 'draft' | 'published' | 'archived';
  avg_rating: number;
  total_ratings: number;
  images?: TourImage[];
  schedules?: TourSchedule[];
  created_at: string;
  updated_at: string;
}
