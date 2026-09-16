package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// A3/A4/A7's validation and mapping edge cases, split out of
// review_admin_handler_test.go to keep both files under the 200-line cap —
// same arrangement as user_admin_handler_security_test.go.

func TestUpdateReviewStatusDraftIsUnprocessable(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch, "/api/v1/admin/reviews/"+testReviewID+"/status",
		`{"status":"draft"}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422; body %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateReviewStatusPublishedSucceeds(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch, "/api/v1/admin/reviews/"+testReviewID+"/status",
		`{"status":"published"}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", rec.Code, rec.Body.String())
	}
}

func TestGetReviewMalformedIDIs400(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodGet, "/api/v1/admin/reviews/not-a-uuid", "", testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestListReviewCategoriesReturnsPlainArray(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{categories: []domain.ReviewCategory{{Name: "Place"}}}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodGet, "/api/v1/admin/review-categories", "", testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	if !strings.HasPrefix(strings.TrimSpace(rec.Body.String()), "[") {
		t.Errorf("body = %s, want a bare JSON array", rec.Body.String())
	}
}

func TestDeleteReviewNotFoundIs404(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{deleteErr: repository.ErrReviewNotFound}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodDelete, "/api/v1/admin/reviews/"+testReviewID, "", testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404; body %s", rec.Code, rec.Body.String())
	}
}
