package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrReviewNotFound is returned by a lookup or guarded write that finds no
// matching, non-deleted review row. Modelled on user_repository.go's
// ErrUserNotFound.
var ErrReviewNotFound = errors.New("repository: review not found")

// reviewSortAllow maps F006 A1's sort_by values to ORDER BY literals; "" is
// the default. An unrecognised value falls back to created_at DESC silently
// (BR-005), the same ResolveSort path every other admin list endpoint uses.
var reviewSortAllow = map[string]string{
	"":              "created_at",
	"created_at":    "created_at",
	"like_count":    "like_count",
	"comment_count": "comment_count",
	"title":         "title",
}

// ReviewListParams adds F006's status/category filters (A1) to the shared
// list params.
type ReviewListParams struct {
	ListParams
	Status           *string
	ReviewCategoryID *uuid.UUID
}

// ReviewWithAuthor is A2's joined review read: the review row plus its
// author/category display fields, scanned in one query so A2 never needs a
// second round trip for metadata alongside the review itself (Architecture:
// "one query for the review plus one for the whole comment set").
type ReviewWithAuthor struct {
	Review          domain.Review
	AuthorName      string
	AuthorAvatarURL *string
	CategoryName    string
}

// ReviewAdminRepository is F006's data access over reviews and their
// (author, category) joins, plus the flat comment read A2's tree is built
// from. Comment writes live in CommentAdminRepository instead, so neither
// file approaches the 200-line cap.
type ReviewAdminRepository interface {
	// List is A1 (implemented in review_admin_repository_read.go).
	List(ctx context.Context, db DB, p ReviewListParams) ([]domain.ReviewListItem, int64, error)
	// Get is A2's joined review read. ErrReviewNotFound signals missing or
	// soft-deleted.
	Get(ctx context.Context, db DB, id uuid.UUID) (*ReviewWithAuthor, error)
	// ListCommentsForReview is A2's comment read. Deliberately has NO
	// deleted_at predicate: comments.parent_id's self-FK is
	// ON DELETE CASCADE, which fires only on a hard DELETE this app never
	// issues, so a soft-deleted parent's replies remain live rows that must
	// stay visible and moderable (BR-003). Filtering deleted_at here would
	// silently orphan them out of the result set entirely, before
	// comment_thread_builder.go ever gets a chance to redact and re-attach
	// them. Do not "fix" this by adding the filter back.
	ListCommentsForReview(ctx context.Context, db DB, reviewID uuid.UUID) ([]domain.Comment, error)
	// UpdateStatus is A3, guarded to published/hidden by the service before
	// this ever runs. ErrReviewNotFound signals a zero-row RETURNING.
	UpdateStatus(ctx context.Context, db DB, id uuid.UUID, status string) (*domain.Review, error)
	// SoftDelete is A4. ErrReviewNotFound signals a zero-row UPDATE.
	SoftDelete(ctx context.Context, db DB, id uuid.UUID) error
	// ListCategories is A7's read-only lookup, ordered by name.
	ListCategories(ctx context.Context, db DB) ([]domain.ReviewCategory, error)
}

// reviewSelectColumns lists the reviews columns in scanReview's exact order,
// every one bound to the `r` alias. The qualification is load-bearing, not
// decoration: A2's Get joins users and review_categories, and all three
// tables carry id/created_at/updated_at/deleted_at, so an unqualified list
// makes Postgres reject the whole statement with "column reference \"id\" is
// ambiguous" (SQLSTATE 42702). Every statement interpolating this constant
// must therefore alias reviews as `r` — including the single-table UPDATEs,
// which is why UpdateStatus writes `UPDATE reviews AS r`. Keeping one
// qualified constant (rather than a qualified and an unqualified copy) is
// what stops the two lists drifting out of scanReview's column order.
const reviewSelectColumns = `r.id, r.user_id, r.review_category_id, r.title, r.slug, r.content, r.thumbnail_url, ` +
	`r.like_count, r.comment_count, r.status, r.created_at, r.updated_at, r.deleted_at`

type reviewAdminRepository struct{}

// NewReviewAdminRepository returns the pgx-backed ReviewAdminRepository.
func NewReviewAdminRepository() ReviewAdminRepository {
	return reviewAdminRepository{}
}

func (reviewAdminRepository) Get(ctx context.Context, db DB, id uuid.UUID) (*ReviewWithAuthor, error) {
	row := db.QueryRow(ctx,
		`SELECT `+reviewSelectColumns+`,
		        COALESCE(u.full_name, ''), u.avatar_url, COALESCE(c.name, '')
		 FROM reviews r
		 LEFT JOIN users u ON u.id = r.user_id
		 LEFT JOIN review_categories c ON c.id = r.review_category_id
		 WHERE r.id = @id AND r.deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)

	var result ReviewWithAuthor
	err := row.Scan(
		&result.Review.ID, &result.Review.UserID, &result.Review.ReviewCategoryID, &result.Review.Title,
		&result.Review.Slug, &result.Review.Content, &result.Review.ThumbnailURL, &result.Review.LikeCount,
		&result.Review.CommentCount, &result.Review.Status, &result.Review.CreatedAt, &result.Review.UpdatedAt,
		&result.Review.DeletedAt,
		&result.AuthorName, &result.AuthorAvatarURL, &result.CategoryName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning review: %w", err)
	}
	return &result, nil
}

func (reviewAdminRepository) ListCommentsForReview(ctx context.Context, db DB, reviewID uuid.UUID) ([]domain.Comment, error) {
	rows, err := db.Query(ctx,
		`SELECT c.id, c.review_id, c.user_id, c.parent_id, c.content, c.is_hidden,
		        c.created_at, c.updated_at, c.deleted_at,
		        u.id, u.full_name, u.avatar_url
		 FROM comments c
		 JOIN users u ON u.id = c.user_id
		 WHERE c.review_id = @review_id
		 ORDER BY c.created_at ASC, c.id`,
		pgx.NamedArgs{"review_id": reviewID},
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing comments for review: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		var author domain.User
		if err := rows.Scan(
			&c.ID, &c.ReviewID, &c.UserID, &c.ParentID, &c.Content, &c.IsHidden,
			&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
			&author.ID, &author.FullName, &author.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("repository: scanning comment: %w", err)
		}
		c.User = &author
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (reviewAdminRepository) UpdateStatus(ctx context.Context, db DB, id uuid.UUID, status string) (*domain.Review, error) {
	row := db.QueryRow(ctx,
		`UPDATE reviews AS r SET status = @status WHERE r.id = @id AND r.deleted_at IS NULL RETURNING `+reviewSelectColumns,
		pgx.NamedArgs{"id": id, "status": status},
	)
	return scanReview(row)
}

func (reviewAdminRepository) SoftDelete(ctx context.Context, db DB, id uuid.UUID) error {
	var deletedID uuid.UUID
	err := db.QueryRow(ctx,
		`UPDATE reviews SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL RETURNING id`,
		pgx.NamedArgs{"id": id},
	).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrReviewNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: soft-deleting review: %w", err)
	}
	return nil
}

// ListCategories is implemented in review_admin_repository_read.go, keeping
// this file under the 200-line cap.

func scanReview(row pgx.Row) (*domain.Review, error) {
	var r domain.Review
	err := row.Scan(
		&r.ID, &r.UserID, &r.ReviewCategoryID, &r.Title, &r.Slug, &r.Content, &r.ThumbnailURL,
		&r.LikeCount, &r.CommentCount, &r.Status, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrReviewNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning review: %w", err)
	}
	return &r, nil
}
