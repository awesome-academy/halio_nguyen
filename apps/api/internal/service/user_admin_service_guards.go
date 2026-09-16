package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// validatePatch is A3's shape check, run before BR-001/BR-002/BR-003: an
// empty patch or an invalid role value is 422, never reaching the guard
// pipeline or opening a transaction.
func validatePatch(patch repository.UserPatch) error {
	if patch.IsActive == nil && patch.Role == nil {
		msg := "At least one of is_active or role is required."
		return apperror.NewUnprocessable(msg, map[string]string{"patch": msg})
	}
	if patch.Role != nil && *patch.Role != domain.RoleAdmin && *patch.Role != domain.RoleUser {
		msg := "Role must be admin or user."
		return apperror.NewUnprocessable(msg, map[string]string{"role": msg})
	}
	return nil
}

// guardLastAdmin is BR-002, refined per phase-07's Key Insights / R1: the
// spec's literal pseudocode ("target.role == admin and active admins <= 1")
// blocks re-activating or promoting the sole admin. The guard must fire only
// on a mutation that would REMOVE an active admin from the pool — demote,
// deactivate, or delete — never on a promotion, reactivation, or no-op.
//
// LockActiveAdmins takes its row lock inside the caller's transaction
// (R2's concurrency fix): two concurrent demotions of the last two admins
// serialize on that SELECT ... FOR UPDATE, so the second one re-reads a
// shrunk set and trips this guard instead of both succeeding. That
// concurrency guarantee itself cannot be exercised by a mock (no Postgres in
// CI, L5) — it remains a manual/local verification step (Implementation
// Steps #11).
func (s *UserAdminService) guardLastAdmin(ctx context.Context, tx repository.DB, target uuid.UUID, patch repository.UserPatch, isDelete bool) error {
	activeAdmins, err := s.repo.LockActiveAdmins(ctx, tx)
	if err != nil {
		return apperror.NewInternal(err)
	}

	targetIsActiveAdmin := containsID(activeAdmins, target)
	if targetIsActiveAdmin && removesAnActiveAdmin(patch, isDelete) && len(activeAdmins) <= 1 {
		msg := "At least one admin account must stay active."
		return apperror.NewConflict(msg, map[string]string{"role": msg})
	}
	return nil
}

// removesAnActiveAdmin decides whether patch/isDelete would strip the
// target's active-admin status. guardLastAdmin only reaches this after
// confirming the target currently IS an active admin, so a promotion
// (target not yet admin) never trips it regardless of what this returns.
func removesAnActiveAdmin(patch repository.UserPatch, isDelete bool) bool {
	if isDelete {
		return true
	}
	if patch.Role != nil && *patch.Role == domain.RoleUser {
		return true
	}
	if patch.IsActive != nil && !*patch.IsActive {
		return true
	}
	return false
}

func containsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func mapUserRepoError(err error) error {
	if errors.Is(err, repository.ErrUserNotFound) {
		return apperror.NewNotFound("User not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
