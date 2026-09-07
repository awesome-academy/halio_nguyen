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
