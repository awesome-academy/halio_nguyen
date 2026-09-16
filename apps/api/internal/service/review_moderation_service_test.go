package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

func TestReviewModerationService_UpdateStatus_DraftIsUnprocessableBeforeAnyWrite(t *testing.T) {
	repo := &mockReviewAdminRepo{}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ReviewStatusDraft)

	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeUnprocessable {
		t.Fatalf("err = %v, want 422 unprocessable", err)
	}
	if repo.updateStatusCalls != 0 {
		t.Errorf("UpdateStatus called %d times, want 0 (validated before any query)", repo.updateStatusCalls)
	}
}

func TestReviewModerationService_UpdateStatus_PublishedSucceeds(t *testing.T) {
	want := &domain.Review{ID: uuid.New(), Status: domain.ReviewStatusPublished}
	repo := &mockReviewAdminRepo{updateStatusResult: want}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	got, err := svc.UpdateStatus(context.Background(), uuid.New(), want.ID, domain.ReviewStatusPublished)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if repo.updateStatusCalls != 1 {
		t.Errorf("UpdateStatus called %d times, want 1", repo.updateStatusCalls)
	}
}

func TestReviewModerationService_UpdateStatus_HiddenSucceeds(t *testing.T) {
	repo := &mockReviewAdminRepo{updateStatusResult: &domain.Review{Status: domain.ReviewStatusHidden}}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	if _, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ReviewStatusHidden); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReviewModerationService_UpdateStatus_NotFoundMapsTo404(t *testing.T) {
	repo := &mockReviewAdminRepo{updateStatusErr: repository.ErrReviewNotFound}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	_, err := svc.UpdateStatus(context.Background(), uuid.New(), uuid.New(), domain.ReviewStatusHidden)
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want 404 not found", err)
	}
}

func TestReviewModerationService_Delete_NotFoundMapsTo404(t *testing.T) {
	repo := &mockReviewAdminRepo{deleteErr: repository.ErrReviewNotFound}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want 404 not found", err)
	}
}

func TestReviewModerationService_Delete_Succeeds(t *testing.T) {
	repo := &mockReviewAdminRepo{}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	if err := svc.Delete(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Errorf("SoftDelete called %d times, want 1", repo.deleteCalls)
	}
}

func TestReviewModerationService_Get_NotFoundMapsTo404(t *testing.T) {
	repo := &mockReviewAdminRepo{getErr: repository.ErrReviewNotFound}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	_, err := svc.Get(context.Background(), uuid.New())
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want 404 not found", err)
	}
}

// The BR-003 case, asserted end to end at the service boundary: a review
// detail composes the repository's flat comment read through
// BuildCommentTree, so a deleted parent with a live reply arrives as a
// placeholder with the reply still attached beneath it.
func TestReviewModerationService_Get_AssemblesCommentTree(t *testing.T) {
	reviewID := uuid.New()
	parentID := uuid.New()
	replyID := uuid.New()
	now := time.Now()
	deletedAt := now.Add(time.Minute)

	repo := &mockReviewAdminRepo{
		getResult: &repository.ReviewWithAuthor{
			Review:       domain.Review{ID: reviewID, Title: "A trip"},
			AuthorName:   "Alice",
			CategoryName: "Place",
		},
		comments: []domain.Comment{
			{ID: parentID, CreatedAt: now, DeletedAt: &deletedAt, Content: "gone"},
			{ID: replyID, ParentID: &parentID, CreatedAt: now.Add(time.Second), Content: "still here"},
		},
	}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	detail, err := svc.Get(context.Background(), reviewID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.AuthorName != "Alice" || detail.CategoryName != "Place" {
		t.Errorf("display fields not composed onto ReviewDetail: %+v", detail)
	}
	if len(detail.Comments) != 1 || !detail.Comments[0].IsDeleted {
		t.Fatalf("want 1 placeholder root, got %+v", detail.Comments)
	}
	if len(detail.Comments[0].Replies) != 1 {
		t.Fatalf("want the live reply preserved beneath the placeholder, got %+v", detail.Comments[0])
	}
}

func TestReviewModerationService_List_ReturnsPaginatedEnvelope(t *testing.T) {
	repo := &mockReviewAdminRepo{listItems: []domain.ReviewListItem{{ID: uuid.New()}}, listTotal: 1}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	page, err := svc.List(context.Background(), repository.ReviewListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Errorf("page = %+v, want 1 item, total 1", page)
	}
}

func TestReviewModerationService_ListCategories_ReturnsPlainSlice(t *testing.T) {
	repo := &mockReviewAdminRepo{categories: []domain.ReviewCategory{{Name: "Place"}}}
	svc := NewReviewModerationService(&fakeDB{}, repo)

	categories, err := svc.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(categories) != 1 {
		t.Errorf("got %d categories, want 1", len(categories))
	}
}
