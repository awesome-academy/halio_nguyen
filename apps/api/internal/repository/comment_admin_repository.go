package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrCommentNotFound is returned by a guarded comment write that finds no
// matching row scoped to both the comment id and its review id.
var ErrCommentNotFound = errors.New("repository: comment not found")

const commentSelectColumns = `id, review_id, user_id, parent_id, content, is_hidden, created_at, updated_at, deleted_at`

// CommentAdminRepository is A5/A6's data access, kept in its own file
// (separate from ReviewAdminRepository) so neither approaches the 200-line
// cap. Every method carries `AND review_id = @review_id` in its WHERE — the
// IDOR control that stops a comment id from being moderated through a
// review it does not belong to (R6, Security Considerations); a mismatched
// pair returns the same ErrCommentNotFound as a missing comment; there is no
// SQL path in this file that could ever touch reviews.like_count or
// reviews.comment_count (BR-002/BR-004).
type CommentAdminRepository interface {
	// UpdateVisibility is A5. ErrCommentNotFound signals a zero-row
	// RETURNING (missing, soft-deleted, or scoped to a different review).
	UpdateVisibility(ctx context.Context, db DB, reviewID, commentID uuid.UUID, isHidden bool) (*domain.Comment, error)
	// SoftDelete is A6. ErrCommentNotFound signals a zero-row UPDATE.
	SoftDelete(ctx context.Context, db DB, reviewID, commentID uuid.UUID) error
}

type commentAdminRepository struct{}

// NewCommentAdminRepository returns the pgx-backed CommentAdminRepository.
func NewCommentAdminRepository() CommentAdminRepository {
	return commentAdminRepository{}
}

func (commentAdminRepository) UpdateVisibility(ctx context.Context, db DB, reviewID, commentID uuid.UUID, isHidden bool) (*domain.Comment, error) {
	row := db.QueryRow(ctx,
		`UPDATE comments SET is_hidden = @is_hidden
		 WHERE id = @id AND review_id = @review_id AND deleted_at IS NULL
		 RETURNING `+commentSelectColumns,
		pgx.NamedArgs{"id": commentID, "review_id": reviewID, "is_hidden": isHidden},
	)

	var c domain.Comment
	err := row.Scan(&c.ID, &c.ReviewID, &c.UserID, &c.ParentID, &c.Content, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCommentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: updating comment visibility: %w", err)
	}
	return &c, nil
}

func (commentAdminRepository) SoftDelete(ctx context.Context, db DB, reviewID, commentID uuid.UUID) error {
	var deletedID uuid.UUID
	err := db.QueryRow(ctx,
		`UPDATE comments SET deleted_at = NOW()
		 WHERE id = @id AND review_id = @review_id AND deleted_at IS NULL
		 RETURNING id`,
		pgx.NamedArgs{"id": commentID, "review_id": reviewID},
	).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCommentNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: soft-deleting comment: %w", err)
	}
	return nil
}
