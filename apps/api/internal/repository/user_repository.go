package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrUserNotFound is returned by a lookup that finds no matching row. The
// service layer maps it to the same anti-enumeration error every other
// login-rejection path uses (BR-001) — it is never surfaced to the client
// as a distinct 404.
var ErrUserNotFound = errors.New("repository: user not found")

// UserRepository is the interface AuthService depends on (A7: interface
// mocks stand in for a real database in tests, since CI has no Postgres —
// L5). Phase 7 extends this same interface with user-management methods.
type UserRepository interface {
	// FindActiveAdminByEmail looks up a non-deleted user by email
	// (case-insensitive). It deliberately does not filter by role or
	// is_active — AuthService checks those so every rejection path (wrong
	// password, wrong role, inactive) costs the same DB + bcrypt work,
	// closing the timing side-channel BR-001/R4 calls out.
	FindActiveAdminByEmail(ctx context.Context, db DB, email string) (*domain.User, error)
	// FindByID looks up a non-deleted user by id, for GET /auth/me.
	FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.User, error)
}

type userRepository struct{}

// NewUserRepository returns the pgx-backed UserRepository implementation.
func NewUserRepository() UserRepository {
	return userRepository{}
}

const userSelectColumns = `id, email, password_hash, full_name, phone, avatar_url, role, is_active, created_at, updated_at, deleted_at`

func (userRepository) FindActiveAdminByEmail(ctx context.Context, db DB, email string) (*domain.User, error) {
	row := db.QueryRow(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE lower(email) = lower(@email) AND deleted_at IS NULL`,
		pgx.NamedArgs{"email": email},
	)
	return scanUser(row)
}

func (userRepository) FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.User, error) {
	row := db.QueryRow(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Phone, &u.AvatarURL,
		&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning user: %w", err)
	}
	return &u, nil
}
