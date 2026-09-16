package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// UserAdminService owns F005's three guards — BR-001 self-lockout, BR-002
// last-admin, BR-003 active-booking delete — plus the plain list/detail
// reads. See phase-07's Architecture for the guard pipeline (BR-001 -> BR-002
// -> BR-003, order load-bearing) and Key Insights for the BR-002 refinement
// (R1): promoting or reactivating the sole admin must succeed, never 409.
type UserAdminService struct {
	db          repository.DB
	repo        repository.UserAdminRepository
	bookingRepo repository.BookingRepository
}

// NewUserAdminService wires the service against the pool (for the guard
// transaction) and its two repositories.
func NewUserAdminService(db repository.DB, repo repository.UserAdminRepository, bookingRepo repository.BookingRepository) *UserAdminService {
	return &UserAdminService{db: db, repo: repo, bookingRepo: bookingRepo}
}

// List is A1.
func (s *UserAdminService) List(ctx context.Context, p repository.UserListParams) (*repository.Paginated[domain.User], error) {
	p.Normalize()
	items, total, err := s.repo.List(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &repository.Paginated[domain.User]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// Get is A2: the profile plus two independent LIFETIME counts — a different
// number from BR-003's active-booking guard count (Key Insights).
func (s *UserAdminService) Get(ctx context.Context, id uuid.UUID) (*domain.UserDetail, error) {
	user, err := s.repo.Get(ctx, s.db, id)
	if err != nil {
		return nil, mapUserRepoError(err)
	}
	bookings, err := s.repo.CountBookings(ctx, s.db, id)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	reviews, err := s.repo.CountReviews(ctx, s.db, id)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &domain.UserDetail{User: user, BookingCount: bookings, ReviewCount: reviews}, nil
}

// Update is A3: the combined status/role PATCH. Validation (empty patch,
// invalid role) runs before BR-001, and BR-001 runs before the transaction
// even opens — so a self-target or a malformed patch writes nothing at all.
func (s *UserAdminService) Update(ctx context.Context, id, actor uuid.UUID, patch repository.UserPatch) (*domain.User, error) {
	if err := validatePatch(patch); err != nil {
		return nil, err
	}
	if id == actor {
		return nil, apperror.NewForbidden("You cannot modify your own account")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.guardLastAdmin(ctx, tx, id, patch, false); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, tx, id, patch)
	if err != nil {
		return nil, mapUserRepoError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return nil, err
	}

	// L1: activity_logs has no CHECK value for this action; slog is the
	// stated fallback trail. Ids and changed-field flags only — never
	// email/phone/name (Security Considerations).
	slog.Info("user updated", "actor_id", actor, "target_id", id,
		"is_active_changed", patch.IsActive != nil, "role_changed", patch.Role != nil)
	return updated, nil
}

// Delete is A4: soft-delete behind all three guards, in order
// (BR-001 -> BR-002 -> BR-003), the last two sharing one transaction with
// the write so a booking created mid-request cannot slip past BR-003
// (D4's implementation note — bookings.user_id's ON DELETE RESTRICT is
// inert against this soft-delete UPDATE).
func (s *UserAdminService) Delete(ctx context.Context, id, actor uuid.UUID) error {
	if id == actor {
		return apperror.NewForbidden("You cannot modify your own account")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.guardLastAdmin(ctx, tx, id, repository.UserPatch{}, true); err != nil {
		return err
	}

	n, err := s.bookingRepo.CountActiveByUser(ctx, tx, id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if n > 0 {
		msg := fmt.Sprintf("Cannot delete: %d active booking(s) reference this account", n)
		return apperror.NewConflict(msg, map[string]string{"bookings": msg})
	}

	if err := s.repo.SoftDelete(ctx, tx, id); err != nil {
		return mapUserRepoError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return err
	}

	slog.Info("user soft-deleted", "actor_id", actor, "target_id", id)
	return nil
}
