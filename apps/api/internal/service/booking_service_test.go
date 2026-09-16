package service

import (
	"context"
	"strings"
	"testing"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// SM-001's transition guards (A3/A5). A4's own rules live in
// booking_cancel_service_test.go; the shared mocks in
// booking_service_mocks_test.go.

func TestIllegalTransitionsAreConflictWithNoWrite(t *testing.T) {
	cases := []struct {
		name    string
		from    string
		action  func(*BookingService) error
		cancels bool
	}{
		{"confirm a confirmed booking", domain.BookingStatusConfirmed, confirmAction, false},
		{"confirm a cancelled booking", domain.BookingStatusCancelled, confirmAction, false},
		{"confirm a completed booking", domain.BookingStatusCompleted, confirmAction, false},
		{"complete a pending booking", domain.BookingStatusPending, completeAction, false},
		{"complete a cancelled booking", domain.BookingStatusCancelled, completeAction, false},
		{"cancel a cancelled booking", domain.BookingStatusCancelled, cancelAction, true},
		{"cancel a completed booking", domain.BookingStatusCompleted, cancelAction, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, schedules, activity, tx := newBookingSvc(tc.from)

			err := tc.action(svc)
			if got := appErrCode(t, err); got != apperror.CodeConflict {
				t.Fatalf("code %q, want conflict", got)
			}
			if !strings.Contains(err.Error(), tc.from) {
				t.Errorf("message %q should name the current status %q", err.Error(), tc.from)
			}
			// SC-002/SC-004: a rejected transition writes nothing at all.
			if schedules.restoreCall != 0 {
				t.Errorf("restored slots on a rejected transition")
			}
			if len(activity.logs) != 0 {
				t.Errorf("wrote an activity log on a rejected transition")
			}
			if tc.cancels && tx.committed {
				t.Errorf("committed the cancel transaction after a rejected guard")
			}
		})
	}
}

func TestMissingBookingIsNotFoundNotConflict(t *testing.T) {
	for name, action := range map[string]func(*BookingService) error{
		"confirm": confirmAction, "complete": completeAction, "cancel": cancelAction,
	} {
		t.Run(name, func(t *testing.T) {
			svc, bookings, _, _, _ := newBookingSvc(domain.BookingStatusPending)
			bookings.missing = true

			if got := appErrCode(t, action(svc)); got != apperror.CodeNotFound {
				t.Fatalf("code %q, want not_found", got)
			}
		})
	}
}

// --- B3: confirming an unpaid booking must SUCCEED ----------------------

// The service never reads payments at all — cash and offline bank_transfer
// settlement depend on it. A 409 here would be a defect, not extra safety
// (R2), so this test guards against someone "hardening" Confirm later.
func TestConfirmUnpaidBookingSucceeds(t *testing.T) {
	svc, bookings, _, _, _ := newBookingSvc(domain.BookingStatusPending)

	confirmed, err := svc.Confirm(context.Background(), testBookingID, testAdminID)
	if err != nil {
		t.Fatalf("confirm on an unpaid pending booking must succeed, got %v", err)
	}
	if confirmed.Status != domain.BookingStatusConfirmed {
		t.Errorf("status %q, want confirmed", confirmed.Status)
	}
	if bookings.lastNext != domain.BookingStatusConfirmed {
		t.Errorf("guarded update wrote %q", bookings.lastNext)
	}
}

func TestCompleteFromConfirmedSucceeds(t *testing.T) {
	svc, _, _, activity, _ := newBookingSvc(domain.BookingStatusConfirmed)

	completed, err := svc.Complete(context.Background(), testBookingID, testAdminID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if completed.Status != domain.BookingStatusCompleted {
		t.Errorf("status %q, want completed", completed.Status)
	}
	// L1: activity_logs' CHECK enum has no value for complete, so the row is
	// deliberately absent and slog carries the trail instead.
	if len(activity.logs) != 0 {
		t.Errorf("complete must not write activity_logs (L1), got %d rows", len(activity.logs))
	}
}
