package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Mocks and fixtures shared by user_admin_service_test.go, split out to keep
// both files under the 200-line cap — same arrangement as
// booking_service_mocks_test.go.

// mockUserAdminRepo models the three call sites the guard pipeline touches.
// lockedAdmins is the BR-002 fixture: every LockActiveAdmins call returns
// this fixed set, standing in for the FOR UPDATE read — the concurrency
// guarantee itself cannot be exercised by a mock (L5, no Postgres in CI).
type mockUserAdminRepo struct {
	repository.UserAdminRepository
	getErr        error
	bookingCount  int64
	reviewCount   int64
	lockedAdmins  []uuid.UUID
	updateErr     error
	updateCalls   int
	softDeleteErr error
	deleteCalls   int
	lockCalls     int
}

func (m *mockUserAdminRepo) CountBookings(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return m.bookingCount, nil
}

func (m *mockUserAdminRepo) CountReviews(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return m.reviewCount, nil
}

func (m *mockUserAdminRepo) LockActiveAdmins(context.Context, repository.DB) ([]uuid.UUID, error) {
	m.lockCalls++
	return m.lockedAdmins, nil
}

func (m *mockUserAdminRepo) Update(_ context.Context, _ repository.DB, id uuid.UUID, patch repository.UserPatch) (*domain.User, error) {
	m.updateCalls++
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	u := &domain.User{ID: id, Role: domain.RoleUser, IsActive: true}
	if patch.Role != nil {
		u.Role = *patch.Role
	}
	if patch.IsActive != nil {
		u.IsActive = *patch.IsActive
	}
	return u, nil
}

func (m *mockUserAdminRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error {
	m.deleteCalls++
	return m.softDeleteErr
}

// mockActiveBookingRepo answers only BookingRepository.CountActiveByUser;
// any other call panics loudly via the embedded interface.
type mockActiveBookingRepo struct {
	repository.BookingRepository
	activeCount int64
	countCalls  int
}

func (m *mockActiveBookingRepo) CountActiveByUser(context.Context, repository.DB, uuid.UUID) (int64, error) {
	m.countCalls++
	return m.activeCount, nil
}

func newUserAdminSvc(repo *mockUserAdminRepo, bookings *mockActiveBookingRepo) (*UserAdminService, *fakeTx) {
	tx := &fakeTx{}
	return NewUserAdminService(&fakeDB{tx: tx}, repo, bookings), tx
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
