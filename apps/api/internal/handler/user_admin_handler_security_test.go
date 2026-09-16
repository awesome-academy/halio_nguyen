package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// password_hash must never appear in A1/A2/A3's raw response bytes, plus the
// PATCH validation edge cases (empty body, invalid role, malformed JSON).
// Scaffolding lives in user_admin_handler_test.go.

func assertNoPasswordHash(t *testing.T, body []byte, hash string) {
	t.Helper()
	s := string(body)
	if strings.Contains(s, "password_hash") || strings.Contains(s, hash) {
		t.Errorf("response leaked password_hash: %s", s)
	}
}

func TestListUsersResponseNeverLeaksPasswordHash(t *testing.T) {
	hash := "s3cr3t-hash-value"
	repo := &stubUserAdminRepo{
		items: []domain.User{{ID: uuid.New(), Email: "a@example.com", PasswordHash: &hash}},
		total: 1,
	}
	e := userAdminGate(t, repo, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, userAdminRequest(t, http.MethodGet, "/api/v1/admin/users", "", testUserAdminActorID, "admin"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	assertNoPasswordHash(t, rec.Body.Bytes(), hash)
}

func TestGetUserResponseNeverLeaksPasswordHash(t *testing.T) {
	hash := "s3cr3t-hash-value"
	repo := &stubUserAdminRepo{
		detail: &domain.User{ID: uuid.MustParse(testUserAdminTargetID), Email: "a@example.com", PasswordHash: &hash},
	}
	e := userAdminGate(t, repo, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, userAdminRequest(t, http.MethodGet, "/api/v1/admin/users/"+testUserAdminTargetID, "", testUserAdminActorID, "admin"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	assertNoPasswordHash(t, rec.Body.Bytes(), hash)
}

func TestUpdateUserResponseNeverLeaksPasswordHash(t *testing.T) {
	hash := "s3cr3t-hash-value"
	repo := &stubUserAdminRepo{
		updated: &domain.User{ID: uuid.MustParse(testUserAdminTargetID), Email: "a@example.com", Role: domain.RoleAdmin, PasswordHash: &hash},
	}
	e := userAdminGate(t, repo, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	req := userAdminRequest(t, http.MethodPatch, "/api/v1/admin/users/"+testUserAdminTargetID, `{"role":"admin"}`, testUserAdminActorID, "admin")
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	assertNoPasswordHash(t, rec.Body.Bytes(), hash)
}

func TestUpdateEmptyBodyIsUnprocessable(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	req := userAdminRequest(t, http.MethodPatch, "/api/v1/admin/users/"+testUserAdminTargetID, `{}`, testUserAdminActorID, "admin")
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422; body %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateInvalidRoleIsUnprocessable(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	req := userAdminRequest(t, http.MethodPatch, "/api/v1/admin/users/"+testUserAdminTargetID, `{"role":"root"}`, testUserAdminActorID, "admin")
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422; body %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateMalformedBodyIs400(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	req := userAdminRequest(t, http.MethodPatch, "/api/v1/admin/users/"+testUserAdminTargetID, `{not json`, testUserAdminActorID, "admin")
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

// pathUUID's 400 is a deliberate deviation from the phase todo's 422
// (documented in user_admin_handler.go's userActionContext) — one
// id-parsing path across the whole portal.
func TestGetUserMalformedIDIs400(t *testing.T) {
	e := userAdminGate(t, &stubUserAdminRepo{}, stubActiveBookingRepo{})
	rec := httptest.NewRecorder()
	req := userAdminRequest(t, http.MethodGet, "/api/v1/admin/users/not-a-uuid", "", testUserAdminActorID, "admin")
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}
