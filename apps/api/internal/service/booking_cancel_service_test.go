package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// A4's own rules: BR-004's required reason and BR-005's slot restore, plus
// the all-or-nothing guarantee over the three writes. Mocks and fixtures live
// in booking_service_mocks_test.go.

func TestCancelRestoresExactlyNumParticipantsAndLogs(t *testing.T) {
	svc, _, schedules, activity, tx := newBookingSvc(domain.BookingStatusConfirmed)

	cancelled, err := svc.Cancel(context.Background(), testBookingID, cancelInput("Customer changed plans"))
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != domain.BookingStatusCancelled {
		t.Errorf("status %q, want cancelled", cancelled.Status)
	}
	// SC-003: exactly num_participants, restored exactly once.
	if schedules.restoreCall != 1 || schedules.restoredBy != 2 {
		t.Errorf("restored %d slots over %d calls, want 2 over 1", schedules.restoredBy, schedules.restoreCall)
	}
	if len(activity.logs) != 1 {
		t.Fatalf("activity rows %d, want 1", len(activity.logs))
	}
	log := activity.logs[0]
	if log.Action != domain.ActionCancelTour {
		t.Errorf("action %q, want cancel_tour", log.Action)
	}
	// The actor is the admin from the JWT, never the booking's customer.
	if log.UserID != testAdminID {
		t.Errorf("user_id %v, want the acting admin %v", log.UserID, testAdminID)
	}
	if log.EntityType != domain.EntityTypeBooking || log.EntityID != testBookingID {
		t.Errorf("entity %s/%v, want booking/%v", log.EntityType, log.EntityID, testBookingID)
	}
	if !tx.committed {
		t.Errorf("cancel did not commit")
	}
}

func TestCancelReasonIsRequiredAndCapped(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"whitespace": "   \t\n ",
		"over cap":   strings.Repeat("x", maxCancellationReasonLen+1),
	}

	for name, reason := range cases {
		t.Run(name, func(t *testing.T) {
			svc, bookings, schedules, activity, tx := newBookingSvc(domain.BookingStatusConfirmed)

			_, err := svc.Cancel(context.Background(), testBookingID, cancelInput(reason))
			appErr := &apperror.Error{}
			if !errors.As(err, &appErr) || appErr.Code != apperror.CodeUnprocessable {
				t.Fatalf("err %v, want unprocessable", err)
			}
			if appErr.StatusCode() != http.StatusUnprocessableEntity {
				t.Errorf("status %d, want 422", appErr.StatusCode())
			}
			if appErr.Fields["cancellation_reason"] == "" {
				t.Errorf("422 should name cancellation_reason, got %v", appErr.Fields)
			}
			// Validation runs before the transaction opens: nothing is written.
			if bookings.cancelCalls != 0 || schedules.restoreCall != 0 || len(activity.logs) != 0 || tx.committed {
				t.Errorf("an invalid reason must write nothing")
			}
		})
	}
}

func TestCancelTrimsTheStoredReason(t *testing.T) {
	svc, bookings, _, _, _ := newBookingSvc(domain.BookingStatusPending)

	if _, err := svc.Cancel(context.Background(), testBookingID, cancelInput("  weather  ")); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if bookings.lastReason != "weather" {
		t.Errorf("stored reason %q, want %q", bookings.lastReason, "weather")
	}
}

// A slot restore or audit insert that fails must take the booking flip down
// with it — a partial commit leaks seats with no reconciliation path (R1).
func TestCancelRollsBackWhenACompanionWriteFails(t *testing.T) {
	t.Run("slot restore fails", func(t *testing.T) {
		svc, _, schedules, _, tx := newBookingSvc(domain.BookingStatusConfirmed)
		schedules.err = errors.New("boom")

		if _, err := svc.Cancel(context.Background(), testBookingID, cancelInput("reason")); err == nil {
			t.Fatal("want an error")
		}
		if tx.committed || !tx.rolledBack {
			t.Errorf("committed=%v rolledBack=%v, want a rollback", tx.committed, tx.rolledBack)
		}
	})

	t.Run("activity insert fails", func(t *testing.T) {
		svc, _, _, activity, tx := newBookingSvc(domain.BookingStatusConfirmed)
		activity.err = errors.New("boom")

		if _, err := svc.Cancel(context.Background(), testBookingID, cancelInput("reason")); err == nil {
			t.Fatal("want an error")
		}
		if tx.committed || !tx.rolledBack {
			t.Errorf("committed=%v rolledBack=%v, want a rollback", tx.committed, tx.rolledBack)
		}
	})
}
