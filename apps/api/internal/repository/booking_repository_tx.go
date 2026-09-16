package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// A3/A4/A5's guarded writes. Each is a single UPDATE ... WHERE status =
// <expected> ... RETURNING: the WHERE clause is the concurrency control, not
// a preceding SELECT. A zero-row result *is* the conflict signal (R3), so a
// simultaneous confirm and cancel can never both win. Cancel's companion
// writes (slot restore, activity log) run against the same DB handle, which
// is the open pgx.Tx when the service passes one.

// UpdateStatusGuarded is A3 (pending -> confirmed) and A5 (confirmed ->
// completed). It deliberately reads nothing first: a pre-read would lose the
// race the guard closes.
func (bookingRepository) UpdateStatusGuarded(ctx context.Context, db DB, id uuid.UUID, expected []string, next string) (*domain.Booking, error) {
	row := db.QueryRow(ctx,
		`UPDATE bookings AS b
		 SET status = @next, updated_at = NOW()
		 WHERE b.id = @id AND b.deleted_at IS NULL AND b.status = ANY(@expected)
		 RETURNING `+bookingColumns,
		pgx.NamedArgs{"id": id, "next": next, "expected": expected},
	)
	return scanBookingRow(row)
}

// CancelGuarded is A4's first write (BR-004): pending|confirmed -> cancelled,
// stamping cancelled_at and the reason. The returned row carries schedule_id
// and num_participants, which is what the caller restores slots by (BR-005)
// — and because those values come out of the same guarded statement, a
// second concurrent cancel gets zero rows and never restores twice (R4).
func (bookingRepository) CancelGuarded(ctx context.Context, db DB, id uuid.UUID, reason string) (*domain.Booking, error) {
	row := db.QueryRow(ctx,
		`UPDATE bookings AS b
		 SET status = @cancelled, cancelled_at = NOW(), cancellation_reason = @reason, updated_at = NOW()
		 WHERE b.id = @id AND b.deleted_at IS NULL AND b.status = ANY(@expected)
		 RETURNING `+bookingColumns,
		pgx.NamedArgs{
			"id":        id,
			"reason":    reason,
			"cancelled": domain.BookingStatusCancelled,
			"expected":  []string{domain.BookingStatusPending, domain.BookingStatusConfirmed},
		},
	)
	return scanBookingRow(row)
}

// scanBookingRow scans a bare booking row (no joins) from either guarded
// UPDATE ... RETURNING. No matching row means the guard rejected the write;
// the caller decides whether that is a 404 or a 409 via CurrentStatus.
func scanBookingRow(row pgx.Row) (*domain.Booking, error) {
	var b domain.Booking
	err := row.Scan(
		&b.ID, &b.BookingCode, &b.UserID, &b.TourID, &b.ScheduleID, &b.NumParticipants,
		&b.UnitPrice, &b.TotalPrice, &b.ContactName, &b.ContactPhone, &b.ContactEmail, &b.SpecialRequests,
		&b.Status, &b.CancelledAt, &b.CancellationReason, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning booking: %w", err)
	}
	return &b, nil
}
