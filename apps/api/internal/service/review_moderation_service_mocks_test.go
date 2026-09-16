package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Mocks shared by review_moderation_service_test.go, split out to keep both
// files under the 200-line cap — same arrangement as
// user_admin_service_mocks_test.go.

// mockReviewAdminRepo models every ReviewAdminRepository call the
// moderation service makes. Call counters let a test assert exactly which
// write ran (or didn't): the 422 draft-status path, in particular, must
// leave updateStatusCalls at zero (BR-002/BR-004 — validated before any
// query, so an invalid request never reaches SQL, let alone a counter
// column this mock has no method to write in the first place).
type mockReviewAdminRepo struct {
	repository.ReviewAdminRepository
	listItems []domain.ReviewListItem
	listTotal int64

	getResult *repository.ReviewWithAuthor
	getErr    error

	comments    []domain.Comment
	commentsErr error

	updateStatusResult *domain.Review
	updateStatusErr    error
	updateStatusCalls  int

	deleteErr   error
	deleteCalls int

	categories    []domain.ReviewCategory
	categoriesErr error
}

func (m *mockReviewAdminRepo) List(context.Context, repository.DB, repository.ReviewListParams) ([]domain.ReviewListItem, int64, error) {
	return m.listItems, m.listTotal, nil
}

func (m *mockReviewAdminRepo) Get(context.Context, repository.DB, uuid.UUID) (*repository.ReviewWithAuthor, error) {
	return m.getResult, m.getErr
}

func (m *mockReviewAdminRepo) ListCommentsForReview(context.Context, repository.DB, uuid.UUID) ([]domain.Comment, error) {
	return m.comments, m.commentsErr
}

func (m *mockReviewAdminRepo) UpdateStatus(_ context.Context, _ repository.DB, _ uuid.UUID, _ string) (*domain.Review, error) {
	m.updateStatusCalls++
	return m.updateStatusResult, m.updateStatusErr
}

func (m *mockReviewAdminRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error {
	m.deleteCalls++
	return m.deleteErr
}

func (m *mockReviewAdminRepo) ListCategories(context.Context, repository.DB) ([]domain.ReviewCategory, error) {
	return m.categories, m.categoriesErr
}
