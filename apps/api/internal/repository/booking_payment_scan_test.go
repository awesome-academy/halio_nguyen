package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// Regression guard for the A2 detail read against a booking with no payments
// row. The bug this covers: payments declares amount/payment_method/status/
// created_at/updated_at NOT NULL, which made it tempting to scan them into
// bare Go types — but a LEFT JOIN that matches nothing yields NULL in every
// one of those columns, so the scan failed and a perfectly valid booking
// returned 500 ("cannot scan NULL into *float64"). CI has no Postgres
// (decisions.md D-A7 / L5), so the assembly is a pure function and tested here.

func TestNullablePaymentAllNullYieldsNilPayment(t *testing.T) {
	// Exactly what a LEFT JOIN miss produces: every column NULL.
	if got := (nullablePayment{}).toDomain(); got != nil {
		t.Fatalf("a payment-less booking must yield a nil Payment, got %+v", got)
	}
}

func TestNullablePaymentPopulatedRowMapsEveryField(t *testing.T) {
	id, bookingID := uuid.New(), uuid.New()
	amount := 2000000.0
	method, status := "internet_banking", "completed"
	ref, bank, reason := "TXN-1", "Vietcombank", "declined"
	paidAt := time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC)
	createdAt := paidAt.Add(-time.Hour)
	updatedAt := paidAt.Add(time.Hour)

	got := nullablePayment{
		id: &id, bookingID: &bookingID, amount: &amount, paymentMethod: &method,
		transactionRef: &ref, bankName: &bank, status: &status, paidAt: &paidAt,
		failedReason: &reason, createdAt: &createdAt, updatedAt: &updatedAt,
	}.toDomain()

	if got == nil {
		t.Fatal("a populated row must yield a Payment")
	}
	if got.ID != id || got.BookingID != bookingID {
		t.Errorf("ids: got %v/%v, want %v/%v", got.ID, got.BookingID, id, bookingID)
	}
	if got.Amount != amount || got.PaymentMethod != method || got.Status != status {
		t.Errorf("core fields: %v/%q/%q", got.Amount, got.PaymentMethod, got.Status)
	}
	if got.TransactionRef != &ref || got.BankName != &bank || got.PaidAt != &paidAt || got.FailedReason != &reason {
		t.Errorf("nullable fields were not carried through verbatim")
	}
	if !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("timestamps: got %v/%v, want %v/%v", got.CreatedAt, got.UpdatedAt, createdAt, updatedAt)
	}
	// Security: A2 never joins user_bank_accounts, and this nil is the only
	// thing preventing account_number from serialising verbatim.
	if got.UserBankAccount != nil {
		t.Errorf("UserBankAccount must stay nil, got %+v", got.UserBankAccount)
	}
}

// A row present but with its optional columns NULL must still produce a
// Payment — only the id decides existence.
func TestNullablePaymentPresentRowWithNullOptionals(t *testing.T) {
	id := uuid.New()
	amount := 0.0
	method, status := "bank_transfer", "pending"

	got := nullablePayment{id: &id, amount: &amount, paymentMethod: &method, status: &status}.toDomain()

	if got == nil {
		t.Fatal("want a Payment")
	}
	if got.TransactionRef != nil || got.BankName != nil || got.PaidAt != nil || got.FailedReason != nil {
		t.Errorf("unset optional columns must stay nil, got %+v", got)
	}
	if got.Status != status || got.PaymentMethod != method {
		t.Errorf("got %q/%q, want %q/%q", got.Status, got.PaymentMethod, status, method)
	}
}
