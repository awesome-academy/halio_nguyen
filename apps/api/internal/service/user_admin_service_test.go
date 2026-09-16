package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// F005's guard pipeline: BR-001 self-lockout, BR-002 last-admin (refined per
// R1), BR-003 active-booking delete. Mocks live in
// user_admin_service_mocks_test.go.

var (
	testActorID  = uuid.MustParse("55555555-5555-5555-5555-555555555001")
	testTargetID = uuid.MustParse("55555555-5555-5555-5555-555555555002")
)

func TestUpdateSelfTargetIsForbiddenWithNoWrite(t *testing.T) {
	repo := &mockUserAdminRepo{}
	svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	_, err := svc.Update(context.Background(), testActorID, testActorID, repository.UserPatch{Role: strPtr(domain.RoleUser)})
	if got := appErrCode(t, err); got != apperror.CodeForbidden {
		t.Fatalf("code %q, want forbidden", got)
	}
	if repo.lockCalls != 0 || repo.updateCalls != 0 || tx.committed {
		t.Errorf("self-target must write nothing: lockCalls=%d updateCalls=%d committed=%v", repo.lockCalls, repo.updateCalls, tx.committed)
	}
}

func TestDeleteSelfTargetIsForbiddenWithNoWrite(t *testing.T) {
	repo := &mockUserAdminRepo{}
	bookings := &mockActiveBookingRepo{}
	svc, tx := newUserAdminSvc(repo, bookings)

	err := svc.Delete(context.Background(), testActorID, testActorID)
	if got := appErrCode(t, err); got != apperror.CodeForbidden {
		t.Fatalf("code %q, want forbidden", got)
	}
	if repo.lockCalls != 0 || repo.deleteCalls != 0 || bookings.countCalls != 0 || tx.committed {
		t.Errorf("self-target must write nothing")
	}
}

func TestLastAdminDemoteDeactivateDeleteAreConflictWithNoWrite(t *testing.T) {
	cases := []struct {
		name string
		run  func(*UserAdminService) error
	}{
		{"demote", func(s *UserAdminService) error {
			_, err := s.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{Role: strPtr(domain.RoleUser)})
			return err
		}},
		{"deactivate", func(s *UserAdminService) error {
			_, err := s.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{IsActive: boolPtr(false)})
			return err
		}},
		{"delete", func(s *UserAdminService) error {
			return s.Delete(context.Background(), testTargetID, testActorID)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The sole active admin IS the target — the set this LockActiveAdmins
			// fixture returns has exactly one member, and it's the target.
			repo := &mockUserAdminRepo{lockedAdmins: []uuid.UUID{testTargetID}}
			svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

			err := tc.run(svc)
			if got := appErrCode(t, err); got != apperror.CodeConflict {
				t.Fatalf("code %q, want conflict", got)
			}
			if repo.updateCalls != 0 || repo.deleteCalls != 0 || tx.committed || !tx.rolledBack {
				t.Errorf("blocked mutation must write nothing and roll back: update=%d delete=%d committed=%v rolledBack=%v",
					repo.updateCalls, repo.deleteCalls, tx.committed, tx.rolledBack)
			}
		})
	}
}

// R1: the spec's literal pseudocode blocks these two — the refined guard
// (removesAnActiveAdmin) must let them through.
func TestPromotingAUserToAdminWithASoleAdminSucceeds(t *testing.T) {
	otherAdmin := uuid.New() // the sole active admin is someone else, not the target
	repo := &mockUserAdminRepo{lockedAdmins: []uuid.UUID{otherAdmin}}
	svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	updated, err := svc.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{Role: strPtr(domain.RoleAdmin)})
	if err != nil {
		t.Fatalf("promotion must succeed, got %v", err)
	}
	if updated.Role != domain.RoleAdmin || repo.updateCalls != 1 || !tx.committed {
		t.Errorf("promotion did not apply: role=%q updateCalls=%d committed=%v", updated.Role, repo.updateCalls, tx.committed)
	}
}

func TestReactivatingADeactivatedSoleAdminSucceeds(t *testing.T) {
	// The deactivated sole admin is not in the active set at all.
	repo := &mockUserAdminRepo{lockedAdmins: nil}
	svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	updated, err := svc.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{IsActive: boolPtr(true)})
	if err != nil {
		t.Fatalf("reactivation must succeed, got %v", err)
	}
	if !updated.IsActive || repo.updateCalls != 1 || !tx.committed {
		t.Errorf("reactivation did not apply: isActive=%v updateCalls=%d committed=%v", updated.IsActive, repo.updateCalls, tx.committed)
	}
}

func TestDeleteWithActiveBookingIsConflictNamingTheCount(t *testing.T) {
	repo := &mockUserAdminRepo{}
	bookings := &mockActiveBookingRepo{activeCount: 3}
	svc, tx := newUserAdminSvc(repo, bookings)

	err := svc.Delete(context.Background(), testTargetID, testActorID)
	if got := appErrCode(t, err); got != apperror.CodeConflict {
		t.Fatalf("code %q, want conflict", got)
	}
	if !strings.Contains(err.Error(), "3") {
		t.Errorf("message %q should name the blocking count", err.Error())
	}
	if repo.deleteCalls != 0 || tx.committed {
		t.Errorf("blocked delete must write nothing")
	}
}

func TestDeleteWithOnlyInactiveBookingsSucceeds(t *testing.T) {
	repo := &mockUserAdminRepo{}
	bookings := &mockActiveBookingRepo{activeCount: 0}
	svc, tx := newUserAdminSvc(repo, bookings)

	if err := svc.Delete(context.Background(), testTargetID, testActorID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if repo.deleteCalls != 1 || !tx.committed {
		t.Errorf("deleteCalls=%d committed=%v", repo.deleteCalls, tx.committed)
	}
}

func TestUpdateEmptyPatchIsUnprocessable(t *testing.T) {
	repo := &mockUserAdminRepo{}
	svc, _ := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	_, err := svc.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{})
	if got := appErrCode(t, err); got != apperror.CodeUnprocessable {
		t.Fatalf("code %q, want unprocessable", got)
	}
	if repo.lockCalls != 0 {
		t.Errorf("validation must run before the transaction opens")
	}
}

func TestUpdateInvalidRoleIsUnprocessable(t *testing.T) {
	repo := &mockUserAdminRepo{}
	svc, _ := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	_, err := svc.Update(context.Background(), testTargetID, testActorID, repository.UserPatch{Role: strPtr("superadmin")})
	if got := appErrCode(t, err); got != apperror.CodeUnprocessable {
		t.Fatalf("code %q, want unprocessable", got)
	}
	if repo.lockCalls != 0 {
		t.Errorf("validation must run before the transaction opens")
	}
}
