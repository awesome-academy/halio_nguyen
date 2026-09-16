package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// List is A1, split out of user_admin_repository.go to keep both files under
// the 200-line cap — the same arrangement as booking_repository_read.go.
func (userAdminRepository) List(ctx context.Context, db DB, p UserListParams) ([]domain.User, int64, error) {
	p.Normalize()
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDesc
	}
	col, dir := ResolveSort(userSortAllow, p.SortBy, sortDir)

	var search *string
	if p.Search != "" {
		search = &p.Search
	}

	rows, err := db.Query(ctx,
		`SELECT `+userSelectColumns+`, COUNT(*) OVER() AS total_count
		 FROM users
		 WHERE deleted_at IS NULL
		   AND (@role::text IS NULL OR role = @role)
		   AND (@is_active::boolean IS NULL OR is_active = @is_active)
		   AND (@search::text IS NULL
		        OR email ILIKE '%' || @search || '%'
		        OR full_name ILIKE '%' || @search || '%')
		 ORDER BY `+col+` `+dir+`, created_at DESC, id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{
			"role": p.Role, "is_active": p.IsActive, "search": search,
			"limit": p.PageSize, "offset": p.Offset(),
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing users: %w", err)
	}
	defer rows.Close()

	items := make([]domain.User, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.AvatarURL,
			&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning user: %w", err)
		}
		items = append(items, u)
	}
	return items, total, rows.Err()
}
