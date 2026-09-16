package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// UserListParams adds F005's role/is_active filters to the shared list
// params (A1).
type UserListParams struct {
	ListParams
	Role     *string
	IsActive *bool
}

// UserPatch is A3's partial update body: is_active and/or role, at least one
// present. The service enforces "at least one" (422) before this ever
// reaches the repository — Update applies whatever is set via COALESCE.
type UserPatch struct {
	IsActive *bool
	Role     *string
}

// userSortAllow maps sort_by values to ORDER BY literals (FR-001's five
// column allowlist); "" is the default. An unrecognised value falls back to
// created_at DESC silently (decisions.md D6), the same ResolveSort path
// every other admin list endpoint uses.
var userSortAllow = map[string]string{
	"":           "created_at",
	"email":      "email",
	"full_name":  "full_name",
	"role":       "role",
	"is_active":  "is_active",
	"created_at": "created_at",
}

// UserAdminRepository is F005's data access: list/search, the two detail
// counts, the BR-002 admin-lock read, the combined status/role update, and
// soft-delete. Every method takes the shared DB so the service's guard
// pipeline (BR-001/BR-002/BR-003) can enlist them all in one transaction.
type UserAdminRepository interface {
	// List is A1 (implemented in user_admin_repository_read.go).
	List(ctx context.Context, db DB, p UserListParams) ([]domain.User, int64, error)
	// Get is A2's profile read. Returns ErrUserNotFound when missing or
	// soft-deleted.
	Get(ctx context.Context, db DB, id uuid.UUID) (*domain.User, error)
	// CountBookings and CountReviews are A2's LIFETIME history counts —
	// every status, every non-deleted row. This is a DIFFERENT number from
	// BookingRepository.CountActiveByUser's BR-003 guard count (pending/
	// confirmed only): two queries, two meanings, never conflated.
	CountBookings(ctx context.Context, db DB, userID uuid.UUID) (int64, error)
	CountReviews(ctx context.Context, db DB, userID uuid.UUID) (int64, error)
	// LockActiveAdmins is BR-002's race-closing read: every currently
	// active, non-deleted admin, row-locked (FOR UPDATE) inside the
	// caller's transaction. Two concurrent demotions of the last two
	// admins serialize on this SELECT, so the second one re-reads a
	// shrunk set instead of both passing a stale COUNT (Key Insights).
	LockActiveAdmins(ctx context.Context, db DB) ([]uuid.UUID, error)
	// Update applies only the supplied patch fields via COALESCE and
	// returns the updated row. ErrUserNotFound signals a zero-row
	// RETURNING — missing or already soft-deleted.
	Update(ctx context.Context, db DB, id uuid.UUID, patch UserPatch) (*domain.User, error)
	// SoftDelete sets deleted_at. ErrUserNotFound signals a zero-row
	// UPDATE.
	SoftDelete(ctx context.Context, db DB, id uuid.UUID) error
}

type userAdminRepository struct{}

// NewUserAdminRepository returns the pgx-backed UserAdminRepository.
func NewUserAdminRepository() UserAdminRepository {
	return userAdminRepository{}
}

func (userAdminRepository) Get(ctx context.Context, db DB, id uuid.UUID) (*domain.User, error) {
	row := db.QueryRow(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	return scanUser(row)
}

func (userAdminRepository) CountBookings(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE user_id = @user_id AND deleted_at IS NULL`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting user bookings: %w", err)
	}
	return n, nil
}

func (userAdminRepository) CountReviews(ctx context.Context, db DB, userID uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM reviews WHERE user_id = @user_id AND deleted_at IS NULL`,
		pgx.NamedArgs{"user_id": userID},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting user reviews: %w", err)
	}
	return n, nil
}

func (userAdminRepository) LockActiveAdmins(ctx context.Context, db DB) ([]uuid.UUID, error) {
	rows, err := db.Query(ctx,
		`SELECT id FROM users WHERE role = @role AND is_active AND deleted_at IS NULL FOR UPDATE`,
		pgx.NamedArgs{"role": domain.RoleAdmin},
	)
	if err != nil {
		return nil, fmt.Errorf("repository: locking active admins: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("repository: scanning locked admin id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (userAdminRepository) Update(ctx context.Context, db DB, id uuid.UUID, patch UserPatch) (*domain.User, error) {
	row := db.QueryRow(ctx,
		`UPDATE users
		 SET is_active = COALESCE(@is_active, is_active), role = COALESCE(@role, role)
		 WHERE id = @id AND deleted_at IS NULL
		 RETURNING `+userSelectColumns,
		pgx.NamedArgs{"id": id, "is_active": patch.IsActive, "role": patch.Role},
	)
	return scanUser(row)
}

func (userAdminRepository) SoftDelete(ctx context.Context, db DB, id uuid.UUID) error {
	var deletedID uuid.UUID
	err := db.QueryRow(ctx,
		`UPDATE users SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL RETURNING id`,
		pgx.NamedArgs{"id": id},
	).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: soft-deleting user: %w", err)
	}
	return nil
}
