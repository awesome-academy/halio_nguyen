package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// A3 Create, and the BR-001/002/005/007 rejection branches it shares with
// Update. SM-001/BR-012/Get live in tour_service_lifecycle_test.go (split to
// keep both files under the 200-line cap).

func TestCreateRollsBackWholeTransactionOnLastScheduleInsertFailure(t *testing.T) {
	tourRepo := &mockTourRepo{}
	imgRepo := &mockTourImageRepo{}
	schRepo := &mockTourScheduleRepo{insertFailOn: 2} // fails on the 2nd (last) schedule
	svc, tx := newTourSvc(tourRepo, activeCategory(), imgRepo, schRepo, &mockBookingRepo{})

	images := []TourImageInput{{ImageURL: "https://example.com/1.jpg"}, {ImageURL: "https://example.com/2.jpg"}}
	schedules := []TourScheduleInput{
		{DepartureDate: mustDate("2026-01-01"), ReturnDate: mustDate("2026-01-05"), AvailableSlots: 10},
		{DepartureDate: mustDate("2026-02-01"), ReturnDate: mustDate("2026-02-05"), AvailableSlots: 10},
	}

	_, err := svc.Create(context.Background(), validTourInput(), images, schedules)
	if err == nil {
		t.Fatal("expected an error from the failing schedule insert")
	}
	if appErrCode(t, err) != apperror.CodeConflict {
		t.Errorf("want 409, got %v", err)
	}
	if tourRepo.inserted == nil {
		t.Error("tours.Insert should have been attempted")
	}
	if len(imgRepo.inserted) != 2 {
		t.Errorf("both images should have been inserted before the schedule failure, got %d", len(imgRepo.inserted))
	}
	if len(schRepo.inserted) != 1 {
		t.Errorf("only the first schedule should have inserted before the failure, got %d", len(schRepo.inserted))
	}
	if tx.committed || !tx.rolledBack {
		t.Errorf("a failure on the last insert must roll back everything: committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
	}
}

func TestCreateCommitsWhenEverythingSucceeds(t *testing.T) {
	tourRepo := &mockTourRepo{}
	imgRepo := &mockTourImageRepo{}
	schRepo := &mockTourScheduleRepo{}
	svc, tx := newTourSvc(tourRepo, activeCategory(), imgRepo, schRepo, &mockBookingRepo{})

	created, err := svc.Create(context.Background(), validTourInput(), nil, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Status != domain.TourStatusDraft {
		t.Errorf("default status = %q, want draft", created.Status)
	}
	if !tx.committed {
		t.Errorf("expected commit, got committed=%v", tx.committed)
	}
}

func TestCreateDerivesSlugFromTitle(t *testing.T) {
	svc, _ := newTourSvc(&mockTourRepo{}, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validTourInput()
	in.Title = "Đà Nẵng Beach Escape"
	created, err := svc.Create(context.Background(), in, nil, nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Slug != "da-nang-beach-escape" {
		t.Errorf("slug = %q", created.Slug)
	}
}

func TestCreateRejectsDiscountAbovePrice(t *testing.T) {
	svc, _ := newTourSvc(&mockTourRepo{}, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validTourInput()
	discount := in.Price + 1
	in.DiscountPrice = &discount

	_, err := svc.Create(context.Background(), in, nil, nil)
	if appErrCode(t, err) != apperror.CodeUnprocessable {
		t.Fatalf("want 422 (BR-001), got %v", err)
	}
}

func TestCreateRejectsNonPositiveDuration(t *testing.T) {
	svc, _ := newTourSvc(&mockTourRepo{}, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validTourInput()
	in.DurationDays = 0

	_, err := svc.Create(context.Background(), in, nil, nil)
	if appErrCode(t, err) != apperror.CodeUnprocessable {
		t.Fatalf("want 422 (BR-002), got %v", err)
	}
}

func TestUpdateRejectsRatingFieldsPresent(t *testing.T) {
	svc, _ := newTourSvc(&mockTourRepo{found: &domain.Tour{ID: uuid.New(), Status: domain.TourStatusDraft}}, activeCategory(), &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	in := validTourInput()
	rating := 4.5
	in.AvgRating = &rating

	_, err := svc.Update(context.Background(), uuid.New(), in)
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeUnprocessable || appErr.Fields["avg_rating"] == "" {
		t.Fatalf("want 422 with fields.avg_rating (BR-005), got %v", err)
	}
}

func TestCreateRejectsInactiveOrMissingCategory(t *testing.T) {
	inactive := &mockTourCategoryRepo{category: &domain.Category{ID: uuid.New(), IsActive: false}}
	svc, _ := newTourSvc(&mockTourRepo{}, inactive, &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	_, err := svc.Create(context.Background(), validTourInput(), nil, nil)
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeUnprocessable || appErr.Fields["category_id"] == "" {
		t.Fatalf("want 422 with fields.category_id (BR-007, inactive), got %v", err)
	}

	missing := &mockTourCategoryRepo{err: repository.ErrCategoryNotFound}
	svc, _ = newTourSvc(&mockTourRepo{}, missing, &mockTourImageRepo{}, &mockTourScheduleRepo{}, &mockBookingRepo{})
	_, err = svc.Create(context.Background(), validTourInput(), nil, nil)
	if appErrCode(t, err) != apperror.CodeUnprocessable {
		t.Fatalf("want 422 (BR-007, not found), got %v", err)
	}
}
