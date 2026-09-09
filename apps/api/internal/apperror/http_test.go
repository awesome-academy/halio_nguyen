package apperror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func newTestContext(method string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestHandler_TypedError(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		wantStatus int
		wantFields bool
	}{
		{
			name:       "conflict with fields",
			err:        NewConflict("duplicate slug", map[string]string{"slug": "already exists"}),
			wantStatus: http.StatusConflict,
			wantFields: true,
		},
		{
			name:       "unprocessable with fields",
			err:        NewUnprocessable("business rule violated", map[string]string{"category_id": "inactive"}),
			wantStatus: http.StatusUnprocessableEntity,
			wantFields: true,
		},
		{
			name:       "not found without fields",
			err:        NewNotFound("tour not found"),
			wantStatus: http.StatusNotFound,
			wantFields: false,
		},
		{
			name:       "unauthorized without fields",
			err:        NewUnauthorized("invalid credentials"),
			wantStatus: http.StatusUnauthorized,
			wantFields: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newTestContext(http.MethodGet)
			Handler(tt.err, c)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			body := rec.Body.String()
			if !strings.Contains(body, string(tt.err.Code)) {
				t.Errorf("body missing code %q: %s", tt.err.Code, body)
			}
			hasFields := strings.Contains(body, `"fields"`)
			if hasFields != tt.wantFields {
				t.Errorf("body fields presence = %v, want %v: %s", hasFields, tt.wantFields, body)
			}
		})
	}
}

func TestHandler_InternalErrorNeverLeaksWrappedCause(t *testing.T) {
	sensitive := errors.New("pq: password authentication failed for user \"admin\"")
	c, rec := newTestContext(http.MethodGet)

	Handler(NewInternal(sensitive), c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "password authentication failed") {
		t.Fatalf("wrapped cause leaked to client: %s", body)
	}
	if !strings.Contains(body, "an internal error occurred") {
		t.Fatalf("expected generic message, got: %s", body)
	}
}

func TestHandler_EchoHTTPError4xxConverted(t *testing.T) {
	c, rec := newTestContext(http.MethodGet)

	Handler(echo.NewHTTPError(http.StatusBadRequest, "malformed body"), c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "malformed body") {
		t.Fatalf("expected the 4xx message to pass through, got: %s", body)
	}
	if !strings.Contains(body, string(CodeBadRequest)) {
		t.Fatalf("expected code %q in body: %s", CodeBadRequest, body)
	}
}

func TestHandler_EchoHTTPError5xxLoggedNotEchoed(t *testing.T) {
	c, rec := newTestContext(http.MethodGet)

	Handler(echo.NewHTTPError(http.StatusServiceUnavailable, "db connection pool exhausted"), c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (5xx echo errors normalize to 500)", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "db connection pool exhausted") {
		t.Fatalf("5xx echo error message leaked to client: %s", body)
	}
	if !strings.Contains(body, "an internal error occurred") {
		t.Fatalf("expected generic message, got: %s", body)
	}
}

func TestHandler_GenericErrorNeverEchoed(t *testing.T) {
	c, rec := newTestContext(http.MethodGet)

	Handler(errors.New("runtime: index out of range [3] with length 2"), c)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "index out of range") {
		t.Fatalf("raw error text leaked to client: %s", body)
	}
	if !strings.Contains(body, string(CodeInternal)) {
		t.Fatalf("expected code %q in body: %s", CodeInternal, body)
	}
}

func TestHandler_HeadRequestWritesNoBody(t *testing.T) {
	c, rec := newTestContext(http.MethodHead)

	Handler(NewNotFound("tour not found"), c)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD response must have no body, got %d bytes: %s", rec.Body.Len(), rec.Body.String())
	}
}

func TestHandler_CommittedResponseIsNoop(t *testing.T) {
	c, rec := newTestContext(http.MethodGet)
	// Simulate a handler that already wrote (and committed) a response
	// before returning an error — Handler must not attempt a second write.
	if err := c.String(http.StatusOK, "already written"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	Handler(NewInternal(errors.New("too late")), c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status changed after commit: got %d, want 200 (first write wins)", rec.Code)
	}
	if rec.Body.String() != "already written" {
		t.Fatalf("body changed after commit: %s", rec.Body.String())
	}
}

func TestStatusCode_AllCodesMapped(t *testing.T) {
	tests := []struct {
		code Code
		want int
	}{
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodeUnprocessable, http.StatusUnprocessableEntity},
		{CodeRateLimited, http.StatusTooManyRequests},
		{CodeInternal, http.StatusInternalServerError},
		{Code("unknown_future_code"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		e := &Error{Code: tt.code}
		if got := e.StatusCode(); got != tt.want {
			t.Errorf("StatusCode(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}
