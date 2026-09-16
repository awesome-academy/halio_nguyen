import { Tour, TourSchedule } from './tour.types';
import { User, UserBankAccount } from './user.types';

export type BookingStatus = 'pending' | 'confirmed' | 'completed' | 'cancelled';
export type PaymentStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'refunded';

export interface Payment {
  id: string;
  booking_id: string;
  user_bank_account_id?: string;
  user_bank_account?: UserBankAccount;
  amount: number;
  payment_method: 'internet_banking' | 'bank_transfer' | 'vnpay' | 'momo';
  transaction_ref?: string;
  bank_name?: string;
  status: PaymentStatus;
  paid_at?: string;
  created_at: string;
}

export interface Booking {
  id: string;
  booking_code: string;
  user_id: string;
  user?: User;
  tour_id: string;
  tour?: Tour;
  schedule_id: string;
  schedule?: TourSchedule;
  num_participants: number;
  unit_price: number;
  total_price: number;
  contact_name: string;
  contact_phone: string;
  contact_email: string;
  special_requests?: string;
  status: BookingStatus;
  cancelled_at?: string;
  cancellation_reason?: string;
  payment?: Payment;
  created_at: string;
  updated_at: string;
}

// --- Admin portal DTOs (F004) -------------------------------------------
// Mirrors apps/api/internal/domain.BookingListItem and the A2 detail shape.

/** One row of GET /api/v1/admin/bookings (domain.BookingListItem).
 * Deliberately narrower than Booking: FR-201 fixes the visible columns, so
 * contact phone/email and special requests are detail-only. */
export interface BookingListItem {
  id: string;
  booking_code: string;
  user_id: string;
  customer_name: string;
  tour_id: string;
  tour_title: string;
  schedule_id: string;
  departure_date: string;
  num_participants: number;
  total_price: number;
  status: BookingStatus;
  created_at: string;
}

/** GET /api/v1/admin/bookings/:id — user/tour/schedule always present;
 * `payment` is absent whenever the booking has no payments row (BR/R5). */
export interface BookingDetail extends Booking {
  user: User;
  tour: Tour;
  schedule: TourSchedule;
}

/** PATCH /api/v1/admin/bookings/:id/cancel */
export interface BookingCancelPayload {
  cancellation_reason: string;
}
