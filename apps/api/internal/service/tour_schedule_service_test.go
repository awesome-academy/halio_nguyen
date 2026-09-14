package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

func newTourScheduleSvc(tourRepo *mockTourRepo, schRepo *mockTourScheduleRepo, bookingRepo *mockBookingRepo) (*TourScheduleService, *fakeTx) {
	tx := &fakeTx{}
	return NewTourScheduleService(&fakeDB{tx: tx}, tourRepo, schRepo, bookingRepo), tx
}

func validScheduleInput() TourScheduleInput {
	return TourScheduleInput{
		DepartureDate: mustDate("2026-06-01"), ReturnDate: mustDate("2026-06-05"), AvailableSlots: 10,
	}
}

// --- A10/A11 BR-002/003/004 ---

func TestScheduleCreateRejectsNegativeSlots(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	svc, _ := newTourScheduleSvc(tourRepo, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validScheduleInput()
	in.AvailableSlots = -1

	_, err := svc.Create(context.Background(), uuid.New(), in)
	if appErrCode(t, err) != apperror.CodeUnprocessable {
		t.Fatalf("want 422 (BR-002), got %v", err)
	}
}

func TestScheduleCreateRejectsReturnBeforeDeparture(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	svc, _ := newTourScheduleSvc(tourRepo, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validScheduleInput()
	in.ReturnDate = mustDate("2026-05-01") // before departure 2026-06-01

	_, err := svc.Create(context.Background(), uuid.New(), in)
	if appErrCode(t, err) != apperror.CodeUnprocessable {
		t.Fatalf("want 422 (BR-003), got %v", err)
	}
}

func TestScheduleCreateMapsDuplicateDepartureDateTo409(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	schRepo := &mockTourScheduleRepo{insertFailOn: 1}
	svc, _ := newTourScheduleSvc(tourRepo, schRepo, &mockBookingRepo{})

	_, err := svc.Create(context.Background(), uuid.New(), validScheduleInput())
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeConflict || appErr.Fields["departure_date"] == "" {
		t.Fatalf("want 409 with fields.departure_date (BR-004), got %v", err)
	}
}

func TestScheduleCreateOnMissingTourIs404(t *testing.T) {
	tourRepo := &mockTourRepo{findErr: repository.ErrTourNotFound}
	svc, _ := newTourScheduleSvc(tourRepo, &mockTourScheduleRepo{}, &mockBookingRepo{})
	_, err := svc.Create(context.Background(), uuid.New(), validScheduleInput())
	if appErrCode(t, err) != apperror.CodeNotFound {
		t.Fatalf("want 404, got %v", err)
	}
}

// --- A12 SM-002 matrix ---

func TestScheduleUpdateStatusSM002Matrix(t *testing.T) {
	tests := []struct {
		from, to string
		wantOK   bool
	}{
		{domain.ScheduleStatusOpen, domain.ScheduleStatusClosed, true},
		{domain.ScheduleStatusClosed, domain.ScheduleStatusOpen, true},
		{domain.ScheduleStatusOpen, domain.ScheduleStatusCancelled, true},
		{domain.ScheduleStatusClosed, domain.ScheduleStatusCancelled, true},
		{domain.ScheduleStatusCancelled, domain.ScheduleStatusOpen, false},
		{domain.ScheduleStatusCancelled, domain.ScheduleStatusClosed, false},
		{domain.ScheduleStatusOpen, domain.ScheduleStatusOpen, false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
			schRepo := &mockTourScheduleRepo{found: &domain.TourSchedule{ID: uuid.New(), Status: tt.from}}
			svc, _ := newTourScheduleSvc(tourRepo, schRepo, &mockBookingRepo{})

			updated, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), tt.to)
			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				if updated.Status != tt.to {
					t.Errorf("status = %q, want %q", updated.Status, tt.to)
				}
			} else if appErrCode(t, err) != apperror.CodeUnprocessable {
				t.Fatalf("expected 422, got %v", err)
			}
		})
	}
}

func TestScheduleUpdateStatusCancelledIsTerminalWithSpecificMessage(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	schRepo := &mockTourScheduleRepo{found: &domain.TourSchedule{ID: uuid.New(), Status: domain.ScheduleStatusCancelled}}
	svc, _ := newTourScheduleSvc(tourRepo, schRepo, &mockBookingRepo{})

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ScheduleStatusOpen)
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Fields["status"] == "" {
		t.Fatalf("want 422 with fields.status, got %v", err)
	}
}

// --- A13 BR-012 delete guard (D4), mirrors tours/A6 ---

func TestScheduleDeleteBlockedByActiveBookingsWritesNothing(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	schRepo := &mockTourScheduleRepo{}
	svc, tx := newTourScheduleSvc(tourRepo, schRepo, &mockBookingRepo{scheduleCount: 2})

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeConflict {
		t.Fatalf("want 409, got %v", err)
	}
	if schRepo.softDeleteCalls != 0 || tx.committed {
		t.Errorf("blocked delete must write nothing: softDeletes=%d committed=%v", schRepo.softDeleteCalls, tx.committed)
	}
}

func TestScheduleDeleteSucceedsWithZeroActiveBookings(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	schRepo := &mockTourScheduleRepo{}
	svc, tx := newTourScheduleSvc(tourRepo, schRepo, &mockBookingRepo{scheduleCount: 0})

	if err := svc.Delete(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if schRepo.softDeleteCalls != 1 || !tx.committed {
		t.Errorf("softDeletes=%d committed=%v", schRepo.softDeleteCalls, tx.committed)
	}
}

func TestScheduleDeleteOnMissingTourIs404(t *testing.T) {
	tourRepo := &mockTourRepo{findErr: repository.ErrTourNotFound}
	svc, _ := newTourScheduleSvc(tourRepo, &mockTourScheduleRepo{}, &mockBookingRepo{})
	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if appErrCode(t, err) != apperror.CodeNotFound {
		t.Fatalf("want 404, got %v", err)
	}
}
