package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// A5 SM-001 transition matrix, A6's BR-012 delete guard (D4), and A2 Get /
// ALG-001. Create + BR-001/002/005/007 live in tour_service_test.go (split
// to keep both files under the 200-line cap).

func TestUpdateStatusSM001Matrix(t *testing.T) {
	tests := []struct {
		from, to string
		wantOK   bool
	}{
		{domain.TourStatusDraft, domain.TourStatusPublished, true},
		{domain.TourStatusPublished, domain.TourStatusArchived, true},
		{domain.TourStatusArchived, domain.TourStatusPublished, true},
		{domain.TourStatusDraft, domain.TourStatusArchived, false},
		{domain.TourStatusPublished, domain.TourStatusDraft, false},
		{domain.TourStatusArchived, domain.TourStatusDraft, false},
		{domain.TourStatusDraft, domain.TourStatusDraft, false},
	}
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New(), Status: tt.from}}
			svc, _ := newTourSvc(tourRepo, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})

			updated, err := svc.UpdateStatus(context.Background(), uuid.New(), tt.to)
			if tt.wantOK {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				if updated.Status != tt.to {
					t.Errorf("status = %q, want %q", updated.Status, tt.to)
				}
			} else {
				if appErrCode(t, err) != apperror.CodeUnprocessable {
					t.Fatalf("expected 422, got %v", err)
				}
			}
		})
	}
}

func TestDeleteBlockedByActiveBookingsWritesNothing(t *testing.T) {
	tourRepo := &mockTourRepo{}
	svc, tx := newTourSvc(tourRepo, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{tourCount: 3})

	err := svc.Delete(context.Background(), uuid.New())
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeConflict {
		t.Fatalf("want 409, got %v", err)
	}
	if tourRepo.softDeleteCalls != 0 || tx.committed || !tx.rolledBack {
		t.Errorf("blocked delete must write nothing: softDeletes=%d committed=%v rolledBack=%v", tourRepo.softDeleteCalls, tx.committed, tx.rolledBack)
	}
}

func TestDeleteSucceedsWithZeroActiveBookings(t *testing.T) {
	tourRepo := &mockTourRepo{}
	svc, tx := newTourSvc(tourRepo, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{tourCount: 0})

	if err := svc.Delete(context.Background(), uuid.New()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if tourRepo.softDeleteCalls != 1 || !tx.committed {
		t.Errorf("softDeletes=%d committed=%v", tourRepo.softDeleteCalls, tx.committed)
	}
}

func TestDeleteUnknownTourIs404(t *testing.T) {
	tourRepo := &mockTourRepo{lockErr: repository.ErrTourNotFound}
	svc, _ := newTourSvc(tourRepo, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	err := svc.Delete(context.Background(), uuid.New())
	if appErrCode(t, err) != apperror.CodeNotFound {
		t.Fatalf("want 404, got %v", err)
	}
}

func TestGetComposesImagesSchedulesAndEffectivePrice(t *testing.T) {
	discount := 80.0
	override := 50.0
	tour := &domain.Tour{ID: uuid.New(), Price: 100, DiscountPrice: &discount}
	tourRepo := &mockTourRepo{found: tour}
	imgRepo := &mockTourImageRepo{listItems: []domain.TourImage{{ID: uuid.New()}}}
	schRepo := &mockTourScheduleRepo{listItems: []domain.TourSchedule{
		{ID: uuid.New(), PriceOverride: &override},
		{ID: uuid.New()},
	}}
	svc, _ := newTourSvc(tourRepo, activeCategory(), imgRepo, schRepo, &mockBookingRepo{})

	detail, err := svc.Get(context.Background(), tour.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(detail.Images) != 1 {
		t.Errorf("images len = %d, want 1", len(detail.Images))
	}
	if len(detail.Schedules) != 2 {
		t.Fatalf("schedules len = %d, want 2", len(detail.Schedules))
	}
	if detail.Schedules[0].EffectivePrice != override {
		t.Errorf("schedule 0 effective_price = %v, want override %v", detail.Schedules[0].EffectivePrice, override)
	}
	if detail.Schedules[1].EffectivePrice != discount {
		t.Errorf("schedule 1 effective_price = %v, want discount %v", detail.Schedules[1].EffectivePrice, discount)
	}
}
