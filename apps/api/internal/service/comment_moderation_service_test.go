package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// mockCommentAdminRepo models the two calls CommentModerationService makes.
// It exposes no method that could write comment_count/like_count
// (BR-002/BR-004) — there is nothing for a test or a future change to call
// even by accident. lastReviewID/lastCommentID let a test assert the R6/IDOR
// scoping ids reached the repository unchanged.
type mockCommentAdminRepo struct {
	repository.CommentAdminRepository
	updateResult  *domain.Comment
	updateErr     error
	updateCalls   int
	deleteErr     error
	deleteCalls   int
	lastReviewID  uuid.UUID
	lastCommentID uuid.UUID
}

func (m *mockCommentAdminRepo) UpdateVisibility(_ context.Context, _ repository.DB, reviewID, commentID uuid.UUID, _ bool) (*domain.Comment, error) {
	m.updateCalls++
	m.lastReviewID, m.lastCommentID = reviewID, commentID
	return m.updateResult, m.updateErr
}

func (m *mockCommentAdminRepo) SoftDelete(_ context.Context, _ repository.DB, reviewID, commentID uuid.UUID) error {
	m.deleteCalls++
	m.lastReviewID, m.lastCommentID = reviewID, commentID
	return m.deleteErr
}

func TestCommentModerationService_UpdateVisibility_Succeeds(t *testing.T) {
	reviewID, commentID := uuid.New(), uuid.New()
	repo := &mockCommentAdminRepo{updateResult: &domain.Comment{ID: commentID, ReviewID: reviewID, IsHidden: true}}
	svc := NewCommentModerationService(&fakeDB{}, repo)

	got, err := svc.UpdateVisibility(context.Background(), uuid.New(), reviewID, commentID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != commentID {
		t.Errorf("got %v, want id %v", got, commentID)
	}
	if repo.updateCalls != 1 {
		t.Errorf("UpdateVisibility called %d times, want 1", repo.updateCalls)
	}
	if repo.lastReviewID != reviewID || repo.lastCommentID != commentID {
		t.Errorf("scoping ids = (%v, %v), want (%v, %v)", repo.lastReviewID, repo.lastCommentID, reviewID, commentID)
	}
}

func TestCommentModerationService_UpdateVisibility_MismatchedReviewIs404(t *testing.T) {
	repo := &mockCommentAdminRepo{updateErr: repository.ErrCommentNotFound}
	svc := NewCommentModerationService(&fakeDB{}, repo)

	_, err := svc.UpdateVisibility(context.Background(), uuid.New(), uuid.New(), uuid.New(), true)
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want 404 not found", err)
	}
}

func TestCommentModerationService_Delete_Succeeds(t *testing.T) {
	reviewID, commentID := uuid.New(), uuid.New()
	repo := &mockCommentAdminRepo{}
	svc := NewCommentModerationService(&fakeDB{}, repo)

	if err := svc.Delete(context.Background(), uuid.New(), reviewID, commentID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.deleteCalls != 1 {
		t.Errorf("SoftDelete called %d times, want 1", repo.deleteCalls)
	}
	if repo.lastReviewID != reviewID || repo.lastCommentID != commentID {
		t.Errorf("scoping ids = (%v, %v), want (%v, %v)", repo.lastReviewID, repo.lastCommentID, reviewID, commentID)
	}
}

func TestCommentModerationService_Delete_MismatchedReviewIs404(t *testing.T) {
	repo := &mockCommentAdminRepo{deleteErr: repository.ErrCommentNotFound}
	svc := NewCommentModerationService(&fakeDB{}, repo)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New(), uuid.New())
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want 404 not found", err)
	}
}
