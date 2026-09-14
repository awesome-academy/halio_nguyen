package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// TourScheduleInput is A10's create DTO (also reused by A3's nested
// schedules[] and by A11's update, which edits the same four fields).
type TourScheduleInput struct {
	DepartureDate  time.Time
	ReturnDate     time.Time
	AvailableSlots int
	PriceOverride  *float64
}

// scheduleTransitions is SM-002: open<->closed, open->cancelled,
// closed->cancelled; cancelled is terminal (BR-010).
var scheduleTransitions = map[string][]string{
	domain.ScheduleStatusOpen:      {domain.ScheduleStatusClosed, domain.ScheduleStatusCancelled},
	domain.ScheduleStatusClosed:    {domain.ScheduleStatusOpen, domain.ScheduleStatusCancelled},
	domain.ScheduleStatusCancelled: {},
}

// TourScheduleService owns F003's departure-schedule rules (CAP-03):
// BR-002/003/004 field validation, BR-010/SM-002 status transitions, and
// BR-012's delete guard (mirrors TourService.Delete).
type TourScheduleService struct {
	db           repository.DB
	tourRepo     repository.TourRepository
	scheduleRepo repository.TourScheduleRepository
	bookingRepo  repository.BookingRepository
}

// NewTourScheduleService wires the service against the pool and its repos.
func NewTourScheduleService(db repository.DB, tourRepo repository.TourRepository, scheduleRepo repository.TourScheduleRepository, bookingRepo repository.BookingRepository) *TourScheduleService {
	return &TourScheduleService{db: db, tourRepo: tourRepo, scheduleRepo: scheduleRepo, bookingRepo: bookingRepo}
}

// Create is A10.
func (s *TourScheduleService) Create(ctx context.Context, tourID uuid.UUID, in TourScheduleInput) (*domain.TourSchedule, error) {
	if err := s.requireTour(ctx, tourID); err != nil {
		return nil, err
	}
	sch, err := buildTourSchedule(tourID, in)
	if err != nil {
		return nil, err
	}
	created, err := s.scheduleRepo.Insert(ctx, s.db, sch)
	if err != nil {
		return nil, mapTourScheduleRepoError(err)
	}
	return created, nil
}

// Update is A11 — same BR-002/003/004 as A10; status is untouched here (A12
// is the single writer, mirroring tours/A5).
func (s *TourScheduleService) Update(ctx context.Context, tourID, scheduleID uuid.UUID, in TourScheduleInput) (*domain.TourSchedule, error) {
	if err := s.requireTour(ctx, tourID); err != nil {
		return nil, err
	}
	sch, err := buildTourSchedule(tourID, in)
	if err != nil {
		return nil, err
	}
	sch.ID = scheduleID
	updated, err := s.scheduleRepo.Update(ctx, s.db, sch)
	if err != nil {
		return nil, mapTourScheduleRepoError(err)
	}
	return updated, nil
}

// UpdateStatus is A12 (BR-010/SM-002). No transaction — same rationale as
// TourService.UpdateStatus (decisions.md §7 D-8 defers the guarded-UPDATE
// hardening for this class of race).
func (s *TourScheduleService) UpdateStatus(ctx context.Context, tourID, scheduleID uuid.UUID, status string) (*domain.TourSchedule, error) {
	if err := s.requireTour(ctx, tourID); err != nil {
		return nil, err
	}
	cur, err := s.scheduleRepo.FindByID(ctx, s.db, tourID, scheduleID)
	if err != nil {
		return nil, mapTourScheduleRepoError(err)
	}
	if !canTransition(scheduleTransitions, cur.Status, status) {
		msg := "This status change isn't allowed."
		if cur.Status == domain.ScheduleStatusCancelled {
			msg = "This schedule is cancelled and cannot be reopened."
		}
		return nil, apperror.NewUnprocessable(msg, map[string]string{"status": msg})
	}
	updated, err := s.scheduleRepo.UpdateStatus(ctx, s.db, tourID, scheduleID, status)
	if err != nil {
		return nil, mapTourScheduleRepoError(err)
	}
	return updated, nil
}

// Delete is A13 (BR-006/BR-012, D4) — mirrors TourService.Delete's guard.
func (s *TourScheduleService) Delete(ctx context.Context, tourID, scheduleID uuid.UUID) error {
	if err := s.requireTour(ctx, tourID); err != nil {
		return err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.scheduleRepo.LockForUpdate(ctx, tx, tourID, scheduleID); err != nil {
		return mapTourScheduleRepoError(err)
	}
	n, err := s.bookingRepo.CountActiveBySchedule(ctx, tx, scheduleID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if n > 0 {
		return apperror.NewConflict(
			fmt.Sprintf("Cannot delete: %d active booking(s) reference this schedule.", n),
			map[string]string{"booking_count": fmt.Sprintf("%d", n)},
		)
	}
	if err := s.scheduleRepo.SoftDelete(ctx, tx, tourID, scheduleID); err != nil {
		return mapTourScheduleRepoError(err)
	}
	return commit(ctx, tx)
}

func (s *TourScheduleService) requireTour(ctx context.Context, tourID uuid.UUID) error {
	if _, err := s.tourRepo.FindByID(ctx, s.db, tourID); err != nil {
		return mapTourRepoError(err)
	}
	return nil
}

// buildTourSchedule validates a TourScheduleInput (BR-002/003) and shapes it
// into the row to write; status always starts/stays 'open' through this
// path (A12 owns every later transition).
func buildTourSchedule(tourID uuid.UUID, in TourScheduleInput) (*domain.TourSchedule, error) {
	fields := map[string]string{}

	if in.AvailableSlots < 0 {
		fields["available_slots"] = "Available slots must be zero or more."
	}
	if in.ReturnDate.Before(in.DepartureDate) {
		fields["return_date"] = "Return date must be on or after the departure date."
	}
	if in.PriceOverride != nil && *in.PriceOverride < 0 {
		fields["price_override"] = "Price override must be zero or more."
	}

	if len(fields) > 0 {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", fields)
	}

	return &domain.TourSchedule{
		TourID:         tourID,
		DepartureDate:  in.DepartureDate,
		ReturnDate:     in.ReturnDate,
		AvailableSlots: in.AvailableSlots,
		PriceOverride:  in.PriceOverride,
		Status:         domain.ScheduleStatusOpen,
	}, nil
}

func mapTourScheduleRepoError(err error) error {
	if errors.Is(err, repository.ErrTourScheduleNotFound) {
		return apperror.NewNotFound("Schedule not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
