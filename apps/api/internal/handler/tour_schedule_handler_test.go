package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScheduleCreateMalformedDateIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	body := `{"departure_date":"01-06-2026","return_date":"2026-06-05","available_slots":10}`
	req := adminRequest(t, http.MethodPost, "/api/v1/admin/tours/"+testTourID+"/schedules", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestScheduleCreateMalformedBodyIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPost, "/api/v1/admin/tours/"+testTourID+"/schedules", "{not json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestScheduleUpdateStatusRequiresStatus(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	path := "/api/v1/admin/tours/" + testTourID + "/schedules/" + testSchedID + "/status"
	req := adminRequest(t, http.MethodPatch, path, "{}")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestScheduleUpdateMalformedScheduleIDIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	body := `{"departure_date":"2026-06-01","return_date":"2026-06-05","available_slots":10}`
	req := adminRequest(t, http.MethodPut, "/api/v1/admin/tours/"+testTourID+"/schedules/not-a-uuid", body)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}
