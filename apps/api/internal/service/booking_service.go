package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// maxCancellationReasonLen caps the free-text reason (BR-004). The column is
// TEXT with no DB limit, so this is the only bound on it — an unbounded
// admin-supplied string is a storage-abuse vector (Security Considerations).
const maxCancellationReasonLen = 1000

// CancelInput carries A4's body plus the request metadata the activity_logs
// row needs. Actor is the admin from the JWT subject, never the booking's
// customer (L1).
type CancelInput struct {
	Reason    string
	Actor     uuid.UUID
	IPAddress string
	UserAgent string
}

// BookingService owns F004's read and transition rules (SM-001). Bookings
// are never created or edited here — the customer site owns creation, and no
// method on this service writes the payments table (refunds are out of
// scope).
type BookingService struct {
	db           repository.DB
	bookingRepo  repository.BookingRepository
	scheduleRepo repository.TourScheduleRepository
	activityRepo repository.ActivityLogRepository
}

// NewBookingService wires the service against the pool (for A4's
// transaction) and its repositories.
func NewBookingService(
	db repository.DB,
	bookingRepo repository.BookingRepository,
	scheduleRepo repository.TourScheduleRepository,
	activityRepo repository.ActivityLogRepository,
) *BookingService {
	return &BookingService{db: db, bookingRepo: bookingRepo, scheduleRepo: scheduleRepo, activityRepo: activityRepo}
}

// List is A1.
func (s *BookingService) List(ctx context.Context, p repository.BookingListParams) (*repository.Paginated[domain.BookingListItem], error) {
	p.Normalize()
	items, total, err := s.bookingRepo.List(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &repository.Paginated[domain.BookingListItem]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// Get is A2. Payment may legitimately be nil — a booking with no payments
// row is not an error (R5).
func (s *BookingService) Get(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, mapBookingRepoError(err)
	}
	return booking, nil
}

// Confirm is A3: pending -> confirmed.
//
// There is deliberately NO payment check here (B3/BR-002). Cash and offline
// bank_transfer settlement are real, so an unpaid booking must still be
// confirmable; the unpaid warning is the UI's job. Rejecting this with a 409
// would be a defect, not extra safety.
func (s *BookingService) Confirm(ctx context.Context, id uuid.UUID, actor uuid.UUID) (*domain.Booking, error) {
	return s.transition(ctx, id, actor, []string{domain.BookingStatusPending}, domain.BookingStatusConfirmed)
}

// Complete is A5: confirmed -> completed (BR-006).
func (s *BookingService) Complete(ctx context.Context, id uuid.UUID, actor uuid.UUID) (*domain.Booking, error) {
	return s.transition(ctx, id, actor, []string{domain.BookingStatusConfirmed}, domain.BookingStatusCompleted)
}

// transition runs A3/A5's guarded update and maps a zero-row result. No
// transaction is needed: the single guarded statement is atomic on its own.
//
// activity_logs' action CHECK enum only permits cancel_tour, so confirm and
// complete cannot be audited to the database (L1, schema frozen). A
// structured slog line is their stated fallback trail — ids only, never the
// booking's contact details.
func (s *BookingService) transition(ctx context.Context, id, actor uuid.UUID, expected []string, next string) (*domain.Booking, error) {
	updated, err := s.bookingRepo.UpdateStatusGuarded(ctx, s.db, id, expected, next)
	if err != nil {
		return nil, s.resolveGuardFailure(ctx, id, err)
	}
	slog.Info("booking status changed", "actor_id", actor, "booking_id", id, "status", next)
	return updated, nil
}

// Cancel is A4: the one multi-table write in this feature. Three writes
// share one transaction — the booking flip, BR-005's slot restore, and the
// single activity_logs row this portal is permitted to write outside
// login/logout. Any failure rolls back all three; a partial commit would
// leak seats with no reconciliation path (R1).
func (s *BookingService) Cancel(ctx context.Context, id uuid.UUID, in CancelInput) (*domain.Booking, error) {
	reason := strings.TrimSpace(in.Reason)
	switch {
	case reason == "":
		msg := "A cancellation reason is required."
		return nil, apperror.NewUnprocessable(msg, map[string]string{"cancellation_reason": msg})
	case len(reason) > maxCancellationReasonLen:
		msg := "Cancellation reason must be 1000 characters or fewer."
		return nil, apperror.NewUnprocessable(msg, map[string]string{"cancellation_reason": msg})
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	cancelled, err := s.bookingRepo.CancelGuarded(ctx, tx, id, reason)
	if err != nil {
		return nil, s.resolveGuardFailure(ctx, id, err)
	}

	// BR-005: exactly num_participants seats go back, no more. There is no
	// database trigger for this and no other admin write to available_slots.
	if err := s.scheduleRepo.RestoreSlots(ctx, tx, cancelled.ScheduleID, cancelled.NumParticipants); err != nil {
		return nil, apperror.NewInternal(err)
	}

	// Unlike the login/logout trail this insert is NOT best-effort: a
	// cancellation that is not audited must not commit.
	err = s.activityRepo.Insert(ctx, tx, repository.NewActivityLog{
		UserID:     in.Actor,
		Action:     domain.ActionCancelTour,
		EntityType: domain.EntityTypeBooking,
		EntityID:   id,
		IPAddress:  in.IPAddress,
		UserAgent:  in.UserAgent,
	})
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	if err := commit(ctx, tx); err != nil {
		return nil, err
	}
	return cancelled, nil
}

// resolveGuardFailure turns a zero-row guarded write into the right status.
// The follow-up existence read runs ONLY after the guard already rejected
// the write, so it is off the hot path and cannot reintroduce the race
// (Implementation Steps #2). A booking that no longer matches but does exist
// was moved by someone else — that is a 409, not a 404.
func (s *BookingService) resolveGuardFailure(ctx context.Context, id uuid.UUID, cause error) error {
	if !errors.Is(cause, repository.ErrBookingNotFound) {
		return mapBookingRepoError(cause)
	}

	current, err := s.bookingRepo.CurrentStatus(ctx, s.db, id)
	if err != nil {
		return mapBookingRepoError(err)
	}
	msg := "Booking is already " + current + "."
	return apperror.NewConflict(msg, map[string]string{"status": msg})
}

func mapBookingRepoError(err error) error {
	if errors.Is(err, repository.ErrBookingNotFound) {
		return apperror.NewNotFound("Booking not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
