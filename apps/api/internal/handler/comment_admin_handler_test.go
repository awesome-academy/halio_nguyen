package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mismatched reviewId/commentId -> 404 (R6/IDOR) and malformed UUID -> 400,
// against the shared gate scaffolding in review_admin_handler_test.go.

func TestUpdateCommentVisibilityMismatchedReviewIs404(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{mismatch: true})
	req := reviewAdminRequest(t, http.MethodPatch,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/"+testCommentID+"/visibility",
		`{"is_hidden":true}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404; body %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteCommentMismatchedReviewIs404(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{mismatch: true})
	req := reviewAdminRequest(t, http.MethodDelete,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/"+testCommentID,
		"", testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404; body %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateCommentVisibilityMalformedCommentIDIs400(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/not-a-uuid/visibility",
		`{"is_hidden":true}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestUpdateCommentVisibilityMalformedReviewIDIs400(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch,
		"/api/v1/admin/reviews/not-a-uuid/comments/"+testCommentID+"/visibility",
		`{"is_hidden":true}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestUpdateCommentVisibilityMissingIsHiddenIs400(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/"+testCommentID+"/visibility",
		`{}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestUpdateCommentVisibilitySucceeds(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodPatch,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/"+testCommentID+"/visibility",
		`{"is_hidden":true}`, testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200; body %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteCommentSucceeds(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	req := reviewAdminRequest(t, http.MethodDelete,
		"/api/v1/admin/reviews/"+testReviewID+"/comments/"+testCommentID,
		"", testReviewAdminActorID, "admin")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d, want 204; body %s", rec.Code, rec.Body.String())
	}
}
