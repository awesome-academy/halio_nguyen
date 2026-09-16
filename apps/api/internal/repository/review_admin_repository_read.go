package repository

import (
	"fmt"

	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// List is A1, split out of review_admin_repository.go to keep both files
// under the 200-line cap — the same arrangement as
// user_admin_repository_read.go. Its joins are LEFT, and every joined text
// column is COALESCE'd: a review whose author or category row cannot be
// found (defensive — both FKs are NOT NULL today) must not 500 a non-pointer
// Scan target the way an earlier phase's bare LEFT JOIN column did.
func (reviewAdminRepository) List(ctx context.Context, db DB, p ReviewListParams) ([]domain.ReviewListItem, int64, error) {
	p.Normalize()
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDesc
	}
	col, dir := ResolveSort(reviewSortAllow, p.SortBy, sortDir)

	var search *string
	if p.Search != "" {
		search = &p.Search
	}

	rows, err := db.Query(ctx,
		`SELECT r.id, r.title, r.slug, r.status, r.like_count, r.comment_count,
		        r.created_at, r.updated_at, r.user_id, COALESCE(u.full_name, '') AS author_name,
		        u.avatar_url, r.review_category_id, COALESCE(c.name, '') AS category_name,
		        COUNT(*) OVER() AS total_count
		 FROM reviews r
		 LEFT JOIN users u ON u.id = r.user_id
		 LEFT JOIN review_categories c ON c.id = r.review_category_id
		 WHERE r.deleted_at IS NULL
		   AND (@status::text IS NULL OR r.status = @status)
		   AND (@category_id::uuid IS NULL OR r.review_category_id = @category_id)
		   AND (@search::text IS NULL OR r.title ILIKE '%' || @search || '%')
		 ORDER BY `+col+` `+dir+`, r.created_at DESC, r.id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{
			"status": p.Status, "category_id": p.ReviewCategoryID, "search": search,
			"limit": p.PageSize, "offset": p.Offset(),
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing reviews: %w", err)
	}
	defer rows.Close()

	items := make([]domain.ReviewListItem, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var item domain.ReviewListItem
		if err := rows.Scan(
			&item.ID, &item.Title, &item.Slug, &item.Status, &item.LikeCount, &item.CommentCount,
			&item.CreatedAt, &item.UpdatedAt, &item.UserID, &item.AuthorName,
			&item.AuthorAvatarURL, &item.ReviewCategoryID, &item.CategoryName, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning review: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// ListCategories is A7's read-only lookup for the filter select, ordered by
// name. Split here (rather than review_admin_repository.go) for the same
// 200-line reason as List.
func (reviewAdminRepository) ListCategories(ctx context.Context, db DB) ([]domain.ReviewCategory, error) {
	rows, err := db.Query(ctx,
		`SELECT id, name, slug, description, created_at, updated_at, deleted_at
		 FROM review_categories WHERE deleted_at IS NULL ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing review categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.ReviewCategory, 0)
	for rows.Next() {
		var rc domain.ReviewCategory
		if err := rows.Scan(&rc.ID, &rc.Name, &rc.Slug, &rc.Description, &rc.CreatedAt, &rc.UpdatedAt, &rc.DeletedAt); err != nil {
			return nil, fmt.Errorf("repository: scanning review category: %w", err)
		}
		categories = append(categories, rc)
	}
	return categories, rows.Err()
}
