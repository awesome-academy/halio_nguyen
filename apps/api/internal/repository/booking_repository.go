package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// BookingRepository is created here, minimal: only the two counts BR-012
// needs (phase-04 Key Insight). Phase 6 extends this same file with the
// booking CRUD it owns — do not duplicate this count logic there.
type BookingRepository interface {
	// CountActiveByTour counts bookings in 'pending'/'confirmed' status
	// against a tour (A6's delete guard).
	CountActiveByTour(ctx context.Context, db DB, tourID uuid.UUID) (int64, error)
	// CountActiveBySchedule counts bookings in 'pending'/'confirmed' status
	// against a schedule (A13's delete guard).
	CountActiveBySchedule(ctx context.Context, db DB, scheduleID uuid.UUID) (int64, error)
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
