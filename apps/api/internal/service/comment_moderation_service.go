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

// CommentModerationService is F006's A5/A6: hide/unhide and soft-delete one
// comment, both scoped to the review it belongs to (R6/IDOR — Security
// Considerations). Neither write ever touches reviews.comment_count; that
// column is trigger-maintained by trg_comments_stat (BR-002/BR-004) — this
// service's repository has no method that could write it even by mistake.
type CommentModerationService struct {
	db   repository.DB
	repo repository.CommentAdminRepository
}

// NewCommentModerationService wires the service against the pool and its
// repository.
func NewCommentModerationService(db repository.DB, repo repository.CommentAdminRepository) *CommentModerationService {
	return &CommentModerationService{db: db, repo: repo}
}

// UpdateVisibility is A5.
func (s *CommentModerationService) UpdateVisibility(ctx context.Context, actor, reviewID, commentID uuid.UUID, isHidden bool) (*domain.Comment, error) {
	updated, err := s.repo.UpdateVisibility(ctx, s.db, reviewID, commentID, isHidden)
	if err != nil {
		return nil, mapCommentRepoError(err)
	}

	slog.Info("comment visibility updated", "actor_id", actor, "review_id", reviewID, "comment_id", commentID, "is_hidden", isHidden)
	return updated, nil
}

// Delete is A6.
func (s *CommentModerationService) Delete(ctx context.Context, actor, reviewID, commentID uuid.UUID) error {
	if err := s.repo.SoftDelete(ctx, s.db, reviewID, commentID); err != nil {
		return mapCommentRepoError(err)
	}

	slog.Info("comment soft-deleted", "actor_id", actor, "review_id", reviewID, "comment_id", commentID)
	return nil
}

func mapCommentRepoError(err error) error {
	if errors.Is(err, repository.ErrCommentNotFound) {
		return apperror.NewNotFound("Comment not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
