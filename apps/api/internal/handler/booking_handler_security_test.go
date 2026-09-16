package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// What must never appear in a booking response, plus the payment-less detail
// path. Scaffolding lives in booking_handler_test.go.

func TestGetBookingWithNoPaymentSerialisesWithoutError(t *testing.T) {
	repo := &stubBookingReadRepo{detail: &domain.Booking{
		ID: uuid.MustParse(testBookingID), BookingCode: "SBT-1", Status: domain.BookingStatusPending,
		User: &domain.User{Email: "a@example.com"}, Tour: &domain.Tour{Title: "Sapa"},
		Schedule: &domain.TourSchedule{}, Payment: nil,
	}}
	e := bookingGate(t, repo)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings/"+testBookingID, ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, present := body["payment"]; present {
		t.Errorf("a booking with no payments row must omit payment entirely: %s", rec.Body.String())
	}
}

// Security: domain.Payment.UserBankAccount is a nested pointer whose
// AccountNumber field carries no json:"-" tag, so a populated pointer would
// serialise the full account number verbatim. A2 never joins
// user_bank_accounts and must leave it nil — asserted against the raw
// response bytes, the same shape as phase-07's password_hash check.
func TestBookingResponsesNeverLeakBankAccountNumbers(t *testing.T) {
	paid := &domain.Booking{
		ID: uuid.MustParse(testBookingID), BookingCode: "SBT-1", Status: domain.BookingStatusConfirmed,
		User: &domain.User{Email: "a@example.com"}, Tour: &domain.Tour{Title: "Sapa"}, Schedule: &domain.TourSchedule{},
		Payment: &domain.Payment{ID: uuid.New(), Amount: 1000, Status: domain.PaymentStatusCompleted},
	}
	e := bookingGate(t, &stubBookingReadRepo{items: []domain.BookingListItem{{ID: paid.ID}}, detail: paid})

	for _, path := range []string{"/api/v1/admin/bookings", "/api/v1/admin/bookings/" + testBookingID} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, path, ""))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d body %s", path, rec.Code, rec.Body.String())
		}
		for _, forbidden := range []string{"account_number", "user_bank_account"} {
			if strings.Contains(rec.Body.String(), forbidden) {
				t.Errorf("%s leaked %q: %s", path, forbidden, rec.Body.String())
			}
		}
	}
}

// The list response is fixed to FR-201's column set: contact phone, email and
// special requests are detail-only, and never widened into the list.
func TestListResponseCarriesNoDetailOnlyPII(t *testing.T) {
	repo := &stubBookingReadRepo{items: []domain.BookingListItem{{
		ID: uuid.New(), BookingCode: "SBT-1", CustomerName: "Mai", TourTitle: "Sapa",
	}}}
	e := bookingGate(t, repo)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings", ""))

	for _, forbidden := range []string{"contact_phone", "contact_email", "special_requests"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Errorf("list response widened with %q: %s", forbidden, rec.Body.String())
		}
	}
}
