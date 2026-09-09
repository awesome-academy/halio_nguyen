package apperror

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// errorBody is the API's single JSON error contract:
//
//	{ "error": { "code": "...", "message": "...", "fields": {...} } }
type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    Code              `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Handler is the echo.HTTPErrorHandler wired in main.go. It normalizes every
// error reaching it — a typed *apperror.Error, a framework *echo.HTTPError,
// or anything else — into the API's one JSON error contract. Unexpected
// errors are logged server-side with the request id and never echoed back to
// the client.
func Handler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		writeError(c, appErr.StatusCode(), appErr.Code, appErr.Message, appErr.Fields)
		return
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.Code >= http.StatusInternalServerError {
			logUnhandled(c, err)
			writeError(c, http.StatusInternalServerError, CodeInternal, "an internal error occurred", nil)
			return
		}
		writeError(c, httpErr.Code, codeForStatus(httpErr.Code), messageForHTTPError(httpErr), nil)
		return
	}

	logUnhandled(c, err)
	writeError(c, http.StatusInternalServerError, CodeInternal, "an internal error occurred", nil)
}

func logUnhandled(c echo.Context, err error) {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	slog.Error("unhandled request error",
		"error", err,
		"request_id", requestID,
		"path", c.Request().URL.Path,
		"method", c.Request().Method,
	)
}

func writeError(c echo.Context, status int, code Code, message string, fields map[string]string) {
	body := errorBody{Error: errorPayload{Code: code, Message: message, Fields: fields}}

	var writeErr error
	if c.Request().Method == http.MethodHead {
		writeErr = c.NoContent(status)
	} else {
		writeErr = c.JSON(status, body)
	}
	if writeErr != nil {
		slog.Error("failed to write error response", "error", writeErr)
	}
}

func codeForStatus(status int) Code {
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	case http.StatusUnprocessableEntity:
		return CodeUnprocessable
	case http.StatusTooManyRequests:
		return CodeRateLimited
	default:
		return CodeInternal
	}
}

func messageForHTTPError(httpErr *echo.HTTPError) string {
	if msg, ok := httpErr.Message.(string); ok {
		return msg
	}
	return http.StatusText(httpErr.Code)
}
