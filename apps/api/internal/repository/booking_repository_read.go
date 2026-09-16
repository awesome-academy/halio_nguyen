package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// A1/A2's read queries. Split out of booking_repository.go to keep both
// files under the 200-line cap.

// List is A1: bookings joined to their user, tour and schedule, filtered and
// paginated. The joins are plain INNER joins — user_id/tour_id/schedule_id
// are all NOT NULL with FK RESTRICT, so a booking can never outlive them.
func (bookingRepository) List(ctx context.Context, db DB, p BookingListParams) ([]domain.BookingListItem, int64, error) {
	p.Normalize()
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDesc
	}
	col, dir := ResolveSort(bookingSortAllow, p.SortBy, sortDir)

	var search *string
	if p.Search != "" {
		search = &p.Search
	}

	rows, err := db.Query(ctx,
		`SELECT b.id, b.booking_code, b.user_id, b.contact_name, b.tour_id, t.title,
		        b.schedule_id, s.departure_date, b.num_participants, b.total_price,
		        b.status, b.created_at, COUNT(*) OVER() AS total_count
		 FROM bookings b
		 JOIN users u ON u.id = b.user_id
		 JOIN tours t ON t.id = b.tour_id
		 JOIN tour_schedules s ON s.id = b.schedule_id
		 WHERE b.deleted_at IS NULL
		   AND (@status::text IS NULL OR b.status = @status)
		   AND (@tour_id::uuid IS NULL OR b.tour_id = @tour_id)
		   AND (@schedule_id::uuid IS NULL OR b.schedule_id = @schedule_id)
		   AND (@date_from::timestamptz IS NULL OR b.created_at >= @date_from)
		   AND (@date_to::timestamptz IS NULL OR b.created_at <= @date_to)
		   AND (@search::text IS NULL
		        OR b.booking_code ILIKE '%' || @search || '%'
		        OR b.contact_name ILIKE '%' || @search || '%'
		        OR b.contact_email ILIKE '%' || @search || '%')
		 ORDER BY `+col+` `+dir+`, b.created_at DESC, b.id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{
			"status": p.Status, "tour_id": p.TourID, "schedule_id": p.ScheduleID,
			"date_from": p.DateFrom, "date_to": p.DateTo, "search": search,
			"limit": p.PageSize, "offset": p.Offset(),
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing bookings: %w", err)
	}
	defer rows.Close()

	items := make([]domain.BookingListItem, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var b domain.BookingListItem
		if err := rows.Scan(
			&b.ID, &b.BookingCode, &b.UserID, &b.CustomerName, &b.TourID, &b.TourTitle,
			&b.ScheduleID, &b.DepartureDate, &b.NumParticipants, &b.TotalPrice,
			&b.Status, &b.CreatedAt, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning booking: %w", err)
		}
		items = append(items, b)
	}
	return items, total, rows.Err()
}

// FindByID is A2. The payments join is a LEFT JOIN over a 0-or-1 row, so
// every payment column scans into a nullable and Payment stays nil when no
// row matched (R5). user_bank_accounts is deliberately never joined: the nil
// Payment.UserBankAccount pointer is the only thing keeping account_number
// out of the response (Security Considerations).
func (bookingRepository) FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Booking, error) {
	row := db.QueryRow(ctx,
		`SELECT `+bookingColumns+`,
		        u.id, u.email, u.full_name, u.phone, u.role, u.is_active,
		        t.id, t.title, t.slug, t.destination, t.duration_days, t.duration_nights, t.thumbnail_url,
		        s.id, s.departure_date, s.return_date, s.available_slots, s.status,
		        p.id, p.booking_id, p.amount, p.payment_method, p.transaction_ref,
		        p.bank_name, p.status, p.paid_at, p.failed_reason, p.created_at, p.updated_at
		 FROM bookings b
		 JOIN users u ON u.id = b.user_id
		 JOIN tours t ON t.id = b.tour_id
		 JOIN tour_schedules s ON s.id = b.schedule_id
		 LEFT JOIN payments p ON p.booking_id = b.id
		 WHERE b.id = @id AND b.deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	return scanBookingDetail(row)
}

// nullablePayment is the scan target for the LEFT JOIN'd payment row. EVERY
// field is a pointer, including the ones payments declares NOT NULL: when the
// join matches nothing Postgres yields NULL in all of them, so scanning
// amount/payment_method/status/created_at/updated_at into their bare Go types
// fails the whole row ("cannot scan NULL into *float64") and takes an
// otherwise-valid booking down with it.
type nullablePayment struct {
	id             *uuid.UUID
	bookingID      *uuid.UUID
	amount         *float64
	paymentMethod  *string
	transactionRef *string
	bankName       *string
	status         *string
	paidAt         *time.Time
	failedReason   *string
	createdAt      *time.Time
	updatedAt      *time.Time
}

// toDomain returns nil when the join matched no row — the id is what decides,
// since it is the only column guaranteed non-NULL on a real payment.
//
// UserBankAccount is deliberately left nil: A2 never joins user_bank_accounts,
// and that nil is the only thing keeping account_number out of the response
// (Security Considerations).
func (p nullablePayment) toDomain() *domain.Payment {
	if p.id == nil {
		return nil
	}
	out := &domain.Payment{
		ID:             *p.id,
		Amount:         deref(p.amount),
		PaymentMethod:  deref(p.paymentMethod),
		TransactionRef: p.transactionRef,
		BankName:       p.bankName,
		Status:         deref(p.status),
		PaidAt:         p.paidAt,
		FailedReason:   p.failedReason,
		CreatedAt:      deref(p.createdAt),
		UpdatedAt:      deref(p.updatedAt),
	}
	if p.bookingID != nil {
		out.BookingID = *p.bookingID
	}
	return out
}

func deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

func scanBookingDetail(row pgx.Row) (*domain.Booking, error) {
	var b domain.Booking
	var u domain.User
	var t domain.Tour
	var s domain.TourSchedule
	var pay nullablePayment

	err := row.Scan(
		&b.ID, &b.BookingCode, &b.UserID, &b.TourID, &b.ScheduleID, &b.NumParticipants,
		&b.UnitPrice, &b.TotalPrice, &b.ContactName, &b.ContactPhone, &b.ContactEmail, &b.SpecialRequests,
		&b.Status, &b.CancelledAt, &b.CancellationReason, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.Role, &u.IsActive,
		&t.ID, &t.Title, &t.Slug, &t.Destination, &t.DurationDays, &t.DurationNights, &t.ThumbnailURL,
		&s.ID, &s.DepartureDate, &s.ReturnDate, &s.AvailableSlots, &s.Status,
		&pay.id, &pay.bookingID, &pay.amount, &pay.paymentMethod, &pay.transactionRef,
		&pay.bankName, &pay.status, &pay.paidAt, &pay.failedReason, &pay.createdAt, &pay.updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning booking detail: %w", err)
	}

	s.TourID = t.ID
	b.User, b.Tour, b.Schedule = &u, &t, &s
	b.Payment = pay.toDomain()
	return &b, nil
}
