package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageUpdateMalformedTourIDIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPut, "/api/v1/admin/tours/not-a-uuid/images/"+testImageID, `{"sort_order":1}`)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestImageUpdateMalformedImageIDIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPut, "/api/v1/admin/tours/"+testTourID+"/images/not-a-uuid", `{"sort_order":1}`)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestImageUpdateRequiresSortOrder(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPut, "/api/v1/admin/tours/"+testTourID+"/images/"+testImageID, "{}")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestImageCreateMalformedBodyIs400(t *testing.T) {
	e := tourGate(t, stubTourRepo{})
	req := adminRequest(t, http.MethodPost, "/api/v1/admin/tours/"+testTourID+"/images", "{not json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}
