package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Two combined/no-op patch shapes against a SOLE active admin. Both resolve
// on the short-circuit order inside removesAnActiveAdmin, which the rest of
// the suite only exercises one field at a time — so a reordering of those
// branches would otherwise pass every existing test. Split into its own file
// to keep user_admin_service_test.go under the 200-line cap.

// A patch that demotes AND reactivates in one body must still be refused:
// the role branch wins, because the mutation's net effect is still the
// removal of the last usable admin.
func TestCombinedDemoteAndReactivatePatchOnSoleAdminIsConflict(t *testing.T) {
	repo := &mockUserAdminRepo{lockedAdmins: []uuid.UUID{testTargetID}}
	svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	_, err := svc.Update(context.Background(), testTargetID, testActorID,
		repository.UserPatch{Role: strPtr(domain.RoleUser), IsActive: boolPtr(true)})

	if got := appErrCode(t, err); got != apperror.CodeConflict {
		t.Fatalf("code %q, want conflict", got)
	}
	if repo.updateCalls != 0 || tx.committed || !tx.rolledBack {
		t.Errorf("blocked mutation must write nothing and roll back: update=%d committed=%v rolledBack=%v",
			repo.updateCalls, tx.committed, tx.rolledBack)
	}
}

// Restating what is already true removes no admin from the pool, so the
// guard must not fire — BR-002 protects against reaching zero usable
// admins, not against touching the last one at all.
func TestNoOpRolePatchOnSoleAdminSucceeds(t *testing.T) {
	repo := &mockUserAdminRepo{lockedAdmins: []uuid.UUID{testTargetID}}
	svc, tx := newUserAdminSvc(repo, &mockActiveBookingRepo{})

	updated, err := svc.Update(context.Background(), testTargetID, testActorID,
		repository.UserPatch{Role: strPtr(domain.RoleAdmin)})

	if err != nil {
		t.Fatalf("no-op role patch must succeed, got %v", err)
	}
	if updated.Role != domain.RoleAdmin || repo.updateCalls != 1 || !tx.committed {
		t.Errorf("no-op patch did not apply: role=%q updateCalls=%d committed=%v",
			updated.Role, repo.updateCalls, tx.committed)
	}
}
