package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// UserAdminHandler is F005's HTTP entry point: list/search (A1), detail
// (A2), the combined status/role PATCH (A3), and soft-delete (A4).
type UserAdminHandler struct {
	svc *service.UserAdminService
}

// NewUserAdminHandler builds the handler over its service.
func NewUserAdminHandler(svc *service.UserAdminService) *UserAdminHandler {
	return &UserAdminHandler{svc: svc}
}

// RegisterRoutes attaches A1-A4 to the A0-gated admin group.
func (h *UserAdminHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/users", h.List)
	gated.GET("/users/:id", h.Get)
	gated.PATCH("/users/:id", h.Update)
	gated.DELETE("/users/:id", h.Delete)
}

// userPatchRequest is A3's explicit DTO — never bind straight into
// domain.User (mass assignment: id/timestamps/password_hash are not
// bindable) or expose repository.UserPatch to the wire directly.
type userPatchRequest struct {
	IsActive *bool   `json:"is_active"`
	Role     *string `json:"role"`
}

func (r userPatchRequest) toPatch() repository.UserPatch {
	return repository.UserPatch{IsActive: r.IsActive, Role: r.Role}
}

func (h *UserAdminHandler) List(c echo.Context) error {
	p, err := parseUserListParams(c)
	if err != nil {
		return err
	}
	page, err := h.svc.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, page)
}

func (h *UserAdminHandler) Get(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	detail, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *UserAdminHandler) Update(c echo.Context) error {
	id, actor, err := userActionContext(c)
	if err != nil {
		return err
	}
	var req userPatchRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	updated, err := h.svc.Update(c.Request().Context(), id, actor, req.toPatch())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *UserAdminHandler) Delete(c echo.Context) error {
	id, actor, err := userActionContext(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id, actor); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// userActionContext resolves A3/A4's two inputs: the target id from the
// path — via the shared pathUUID, which returns 400 on a malformed id
// rather than the phase todo's 422 (a deliberate deviation: one id-parsing
// path across the whole admin portal beats a second parser for this one
// route) — and the acting admin from the verified JWT subject. BR-001
// compares the target against this value, never against a header or body
// field the caller could spoof.
func userActionContext(c echo.Context) (id, actor uuid.UUID, err error) {
	id, err = pathUUID(c)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	claims, ok := adminmw.ClaimsFrom(c)
	if !ok {
		return uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	actor, err = uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	return id, actor, nil
}
