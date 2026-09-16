package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// A1's query-string parsing: FR-202's filters and the date-range bounds.
// Scaffolding lives in booking_handler_test.go.

// FR-202: an inverted range is rejected before the query runs, rather than
// silently returning an empty page that reads as "no bookings".
func TestInvertedDateRangeIs422(t *testing.T) {
	repo := &stubBookingReadRepo{}
	e := bookingGate(t, repo)
	req := bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings?date_from=2026-09-10&date_to=2026-09-01", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422; body %s", rec.Code, rec.Body.String())
	}
	if repo.lastPar.DateFrom != nil {
		t.Errorf("the list query ran despite an inverted range")
	}
}

func TestMalformedListFiltersAre400(t *testing.T) {
	cases := map[string]string{
		"bad date_from":   "date_from=10-09-2026",
		"bad date_to":     "date_to=not-a-date",
		"bad tour_id":     "tour_id=not-a-uuid",
		"bad schedule_id": "schedule_id=not-a-uuid",
	}
	e := bookingGate(t, &stubBookingReadRepo{})

	for name, qs := range cases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings?"+qs, ""))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status %d, want 400; body %s", rec.Code, rec.Body.String())
			}
		})
	}
}

// date_to is inclusive of the whole day against a TIMESTAMPTZ column.
func TestDateToCoversTheWholeDay(t *testing.T) {
	repo := &stubBookingReadRepo{}
	e := bookingGate(t, repo)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings?date_to=2026-09-10", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	to := repo.lastPar.DateTo
	if to == nil {
		t.Fatal("date_to was not parsed")
	}
	if to.Year() != 2026 || to.Month() != time.September || to.Day() != 10 || to.Hour() != 23 {
		t.Errorf("date_to = %v, want the end of 2026-09-10", to)
	}
}

func TestListBookingsReturnsPaginatedEnvelope(t *testing.T) {
	repo := &stubBookingReadRepo{items: []domain.BookingListItem{{ID: uuid.New(), BookingCode: "SBT-1", CustomerName: "Mai"}}}
	e := bookingGate(t, repo)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, bookingRequest(t, http.MethodGet, "/api/v1/admin/bookings?page_size=5", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	for _, key := range []string{`"items"`, `"total":1`, `"page":1`, `"page_size":5`} {
		if !strings.Contains(rec.Body.String(), key) {
			t.Errorf("body missing %s: %s", key, rec.Body.String())
		}
	}
}
