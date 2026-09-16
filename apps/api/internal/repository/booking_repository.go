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

// ErrBookingNotFound is returned when no non-deleted booking matches (BR-001).
var ErrBookingNotFound = errors.New("repository: booking not found")

// BookingListParams adds F004's filters to the shared list params. DateFrom
// and DateTo bound bookings.created_at (FR-202's created-date range), not the
// departure date.
type BookingListParams struct {
	ListParams
	Status     *string
	TourID     *uuid.UUID
	ScheduleID *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
}

// bookingSortAllow maps sort_by values to ORDER BY literals (FR-204); "" is
// the default. An unrecognised value falls back to created_at DESC silently
// (SC-001), which is why List defaults SortDir to desc when unset.
var bookingSortAllow = map[string]string{
	"":            "b.created_at",
	"created_at":  "b.created_at",
	"total_price": "b.total_price",
	"status":      "b.status",
}

// bookingColumns is the full booking row, used by the detail read and by
// both guarded UPDATE ... RETURNING statements.
const bookingColumns = `b.id, b.booking_code, b.user_id, b.tour_id, b.schedule_id, b.num_participants,
	b.unit_price, b.total_price, b.contact_name, b.contact_phone, b.contact_email, b.special_requests,
	b.status, b.cancelled_at, b.cancellation_reason, b.created_at, b.updated_at, b.deleted_at`

// BookingRepository is the pgx data access A1-A5 depend on, plus the two
// active-booking counts phase-04's delete guards (BR-012) already use.
// Methods take the shared DB so Cancel's three writes enlist in one
// transaction.
type BookingRepository interface {
	// CountActiveByTour counts bookings in 'pending'/'confirmed' status
	// against a tour (A6's delete guard).
	CountActiveByTour(ctx context.Context, db DB, tourID uuid.UUID) (int64, error)
	// CountActiveBySchedule counts bookings in 'pending'/'confirmed' status
	// against a schedule (A13's delete guard).
	CountActiveBySchedule(ctx context.Context, db DB, scheduleID uuid.UUID) (int64, error)
	// CountActiveByUser counts bookings in 'pending'/'confirmed' status
	// against a user (phase-07's BR-003 delete guard). A third sibling of
	// the two counts above rather than a second repository (Key Insights,
	// R6) — same shape, one different WHERE column.
	CountActiveByUser(ctx context.Context, db DB, userID uuid.UUID) (int64, error)

	// List is A1 (implemented in booking_repository_read.go).
	List(ctx context.Context, db DB, p BookingListParams) ([]domain.BookingListItem, int64, error)
	// FindByID is A2: the booking with its user, tour, schedule and its
	// 0-or-1 payment row. Payment is nil when no payment exists.
	FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Booking, error)
	// CurrentStatus reads just the status of a non-deleted booking. It is
	// the 404-vs-409 discriminator called only after a guarded UPDATE
	// returned zero rows — never before one (that would reintroduce the
	// race the guard exists to close).
	CurrentStatus(ctx context.Context, db DB, id uuid.UUID) (string, error)

	// UpdateStatusGuarded is A3/A5: one UPDATE ... WHERE status = expected
	// ... RETURNING. A zero-row result is the conflict signal and surfaces
	// as ErrBookingNotFound; the caller resolves 404 vs 409 with
	// CurrentStatus (implemented in booking_repository_tx.go).
	UpdateStatusGuarded(ctx context.Context, db DB, id uuid.UUID, expected []string, next string) (*domain.Booking, error)
	// CancelGuarded is A4's first write: pending|confirmed -> cancelled,
	// stamping cancelled_at and the reason, returning the row so the caller
	// can restore exactly num_participants slots (BR-005).
	CancelGuarded(ctx context.Context, db DB, id uuid.UUID, reason string) (*domain.Booking, error)
}

type bookingRepository struct{}

// NewBookingRepository returns the pgx-backed BookingRepository.
func NewBookingRepository() BookingRepository {
	return bookingRepository{}
}

func (bookingRepository) CountActiveByTour(ctx context.Context, db DB, tourID uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings
		 WHERE tour_id = @tour_id AND status IN (@pending, @confirmed) AND deleted_at IS NULL`,
		pgx.NamedArgs{"tour_id": tourID, "pending": domain.BookingStatusPending, "confirmed": domain.BookingStatusConfirmed},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting active bookings for tour: %w", err)
	}
	return n, nil
}

func (bookingRepository) CountActiveBySchedule(ctx context.Context, db DB, scheduleID uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings
		 WHERE schedule_id = @schedule_id AND status IN (@pending, @confirmed) AND deleted_at IS NULL`,
		pgx.NamedArgs{"schedule_id": scheduleID, "pending": domain.BookingStatusPending, "confirmed": domain.BookingStatusConfirmed},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting active bookings for schedule: %w", err)
	}
	return n, nil
}

func (bookingRepository) CountActiveByUser(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings
		 WHERE user_id = @user_id AND status IN (@pending, @confirmed) AND deleted_at IS NULL`,
		pgx.NamedArgs{"user_id": userID, "pending": domain.BookingStatusPending, "confirmed": domain.BookingStatusConfirmed},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting active bookings for user: %w", err)
	}
	return n, nil
}

func (bookingRepository) CurrentStatus(ctx context.Context, db DB, id uuid.UUID) (string, error) {
	var status string
	err := db.QueryRow(ctx,
		`SELECT status FROM bookings WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrBookingNotFound
	}
	if err != nil {
		return "", fmt.Errorf("repository: reading booking status: %w", err)
	}
	return status, nil
}
