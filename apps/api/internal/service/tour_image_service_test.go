package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

func newTourImageSvc(tourRepo *mockTourRepo, imgRepo *mockTourImageRepo) *TourImageService {
	return NewTourImageService(&fakeDB{tx: &fakeTx{}}, tourRepo, imgRepo)
}

func TestImageCreateAssignsNextSortOrder(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	imgRepo := &mockTourImageRepo{nextSortOrder: 3}
	svc := newTourImageSvc(tourRepo, imgRepo)

	created, err := svc.Create(context.Background(), uuid.New(), TourImageInput{ImageURL: "https://example.com/a.jpg"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.SortOrder != 3 {
		t.Errorf("sort_order = %d, want 3", created.SortOrder)
	}
}

func TestImageCreateRejectsNonHTTPURL(t *testing.T) {
	svc := newTourImageSvc(&mockTourRepo{found: &domain.Tour{ID: uuid.New()}}, &mockTourImageRepo{})
	_, err := svc.Create(context.Background(), uuid.New(), TourImageInput{ImageURL: "javascript:alert(1)"})
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeUnprocessable || appErr.Fields["image_url"] == "" {
		t.Fatalf("want 422 with fields.image_url, got %v", err)
	}
}

func TestImageCreateOnMissingTourIs404(t *testing.T) {
	svc := newTourImageSvc(&mockTourRepo{findErr: repository.ErrTourNotFound}, &mockTourImageRepo{})
	_, err := svc.Create(context.Background(), uuid.New(), TourImageInput{ImageURL: "https://example.com/a.jpg"})
	if appErrCode(t, err) != apperror.CodeNotFound {
		t.Fatalf("want 404, got %v", err)
	}
}

func TestImageUpdateTrustsSuppliedSortOrderVerbatim(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	imgRepo := &mockTourImageRepo{}
	svc := newTourImageSvc(tourRepo, imgRepo)

	caption := "New caption"
	updated, err := svc.Update(context.Background(), uuid.New(), uuid.New(), TourImageUpdateInput{Caption: &caption, SortOrder: 7})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.SortOrder != 7 {
		t.Errorf("sort_order = %d, want 7 (BR-011: trusted verbatim)", updated.SortOrder)
	}
}

func TestImageDeleteIsIdempotentOnAlreadyDeleted(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	imgRepo := &mockTourImageRepo{softDeleteErr: repository.ErrTourImageNotFound}
	svc := newTourImageSvc(tourRepo, imgRepo)

	if err := svc.Delete(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("Delete of an already-deleted image must be idempotent (204), got %v", err)
	}
}

func TestImageDeleteSucceeds(t *testing.T) {
	tourRepo := &mockTourRepo{found: &domain.Tour{ID: uuid.New()}}
	imgRepo := &mockTourImageRepo{}
	svc := newTourImageSvc(tourRepo, imgRepo)

	if err := svc.Delete(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if imgRepo.softDeleteCalls != 1 {
		t.Errorf("softDeleteCalls = %d, want 1", imgRepo.softDeleteCalls)
	}
}
