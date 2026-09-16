package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// reviewStatusUpdatable is SM-001's two-value moderation target set: draft
// is deliberately excluded — it belongs to the authoring flow, out of scope
// here, and a PATCH naming it is 422 (validated below, before any query
// runs), never the DB's own three-value CHECK.
var reviewStatusUpdatable = map[string]bool{
	domain.ReviewStatusPublished: true,
	domain.ReviewStatusHidden:    true,
}

// ReviewModerationService is F006's A1/A2/A3/A4/A7: list/search, the review
// plus its full nested comment tree, the published/hidden status flip, and
// soft-delete. Every mutation writes exactly one column and logs an slog
// trail (actor, action, review id) — the only moderation record this
// feature ships (L1, activity_logs has no CHECK value for it).
type ReviewModerationService struct {
	db   repository.DB
	repo repository.ReviewAdminRepository
}

// NewReviewModerationService wires the service against the pool and its
// repository.
func NewReviewModerationService(db repository.DB, repo repository.ReviewAdminRepository) *ReviewModerationService {
	return &ReviewModerationService{db: db, repo: repo}
}

// List is A1.
func (s *ReviewModerationService) List(ctx context.Context, p repository.ReviewListParams) (*repository.Paginated[domain.ReviewListItem], error) {
	p.Normalize()
	items, total, err := s.repo.List(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &repository.Paginated[domain.ReviewListItem]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// Get is A2: the review plus its author/category display fields and the
// assembled comment tree (comment_thread_builder.go). Two repository calls
// total — never N+1 per comment node.
func (s *ReviewModerationService) Get(ctx context.Context, id uuid.UUID) (*domain.ReviewDetail, error) {
	row, err := s.repo.Get(ctx, s.db, id)
	if err != nil {
		return nil, mapReviewRepoError(err)
	}

	comments, err := s.repo.ListCommentsForReview(ctx, s.db, id)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	review := row.Review
	return &domain.ReviewDetail{
		Review:          &review,
		AuthorName:      row.AuthorName,
		AuthorAvatarURL: row.AuthorAvatarURL,
		CategoryName:    row.CategoryName,
		Comments:        BuildCommentTree(comments),
	}, nil
}

// UpdateStatus is A3: published <-> hidden only (SM-001). draft is rejected
// here, before the repository is ever called, so an invalid request writes
// nothing.
func (s *ReviewModerationService) UpdateStatus(ctx context.Context, actor, id uuid.UUID, status string) (*domain.Review, error) {
	if !reviewStatusUpdatable[status] {
		msg := "Status must be published or hidden."
		return nil, apperror.NewUnprocessable(msg, map[string]string{"status": msg})
	}

	updated, err := s.repo.UpdateStatus(ctx, s.db, id, status)
	if err != nil {
		return nil, mapReviewRepoError(err)
	}

	slog.Info("review status updated", "actor_id", actor, "review_id", id, "status", status)
	return updated, nil
}

// Delete is A4: soft-delete, terminal (§5 F006 default — no restore path).
func (s *ReviewModerationService) Delete(ctx context.Context, actor, id uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, s.db, id); err != nil {
		return mapReviewRepoError(err)
	}

	slog.Info("review soft-deleted", "actor_id", actor, "review_id", id)
	return nil
}

// ListCategories is A7's read-only lookup for the filter select.
func (s *ReviewModerationService) ListCategories(ctx context.Context) ([]domain.ReviewCategory, error) {
	categories, err := s.repo.ListCategories(ctx, s.db)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return categories, nil
}

func mapReviewRepoError(err error) error {
	if errors.Is(err, repository.ErrReviewNotFound) {
		return apperror.NewNotFound("Review not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
