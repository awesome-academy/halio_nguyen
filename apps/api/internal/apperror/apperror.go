// Package apperror defines the one typed error every repository, service,
// and handler layer returns, and the single JSON shape it maps to on the
// wire. No layer below the HTTP boundary depends on echo.
package apperror

import "net/http"

// Code is a stable, machine-readable error category. It appears verbatim in
// the JSON error body's "code" field.
type Code string

// The full set of error codes the API ever emits. Every later phase reuses
// these constructors instead of inventing new codes.
const (
	CodeBadRequest    Code = "bad_request"
	CodeUnauthorized  Code = "unauthorized"
	CodeForbidden     Code = "forbidden"
	CodeNotFound      Code = "not_found"
	CodeConflict      Code = "conflict"
	CodeUnprocessable Code = "unprocessable"
	CodeRateLimited   Code = "rate_limited"
	CodeInternal      Code = "internal"
)

// Error is the typed error returned by services and repositories. Fields is
// only populated for 409/422 responses (field-scoped validation/conflict
// messages); Err carries the wrapped cause for logging and is never
// serialized to the client.
type Error struct {
	Code    Code
	Message string
	Fields  map[string]string
	Err     error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap exposes the wrapped cause to errors.Is/errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// StatusCode maps the error's Code to its HTTP status.
func (e *Error) StatusCode() int {
	switch e.Code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// NewBadRequest builds a 400 error for malformed input the handler itself
// can detect (bad query params, malformed JSON body).
func NewBadRequest(message string) *Error {
	return &Error{Code: CodeBadRequest, Message: message}
}

// NewUnauthorized builds a 401 error for a missing or invalid credential.
func NewUnauthorized(message string) *Error {
	return &Error{Code: CodeUnauthorized, Message: message}
}

// NewForbidden builds a 403 error for an authenticated caller lacking the
// required role or ownership.
func NewForbidden(message string) *Error {
	return &Error{Code: CodeForbidden, Message: message}
}

// NewNotFound builds a 404 error for a missing resource.
func NewNotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message}
}

// NewConflict builds a 409 error, optionally naming the fields that caused
// the conflict (e.g. a duplicate slug, a dependent record blocking delete).
func NewConflict(message string, fields map[string]string) *Error {
	return &Error{Code: CodeConflict, Message: message, Fields: fields}
}

// NewUnprocessable builds a 422 error for semantically invalid input that
// passed shape validation (e.g. a business rule violation).
func NewUnprocessable(message string, fields map[string]string) *Error {
	return &Error{Code: CodeUnprocessable, Message: message, Fields: fields}
}

// NewRateLimited builds a 429 error for a throttled request.
func NewRateLimited(message string) *Error {
	return &Error{Code: CodeRateLimited, Message: message}
}

// NewInternal wraps an unexpected error as a 500. The wrapped err is never
// serialized to the client — only logged server-side by apperror.Handler.
func NewInternal(err error) *Error {
	return &Error{Code: CodeInternal, Message: "an internal error occurred", Err: err}
}
