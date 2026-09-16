package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// ReviewAdminHandler is F006's HTTP entry point: list/search (A1), detail
// with the full comment tree (A2), the published/hidden status flip (A3),
// soft-delete (A4), and the read-only category lookup (A7). A5/A6 (comment
// moderation) are registered here too, on a nested group, so router.go
// still gains exactly one RegisterRoutes call.
type ReviewAdminHandler struct {
	svc        *service.ReviewModerationService
	commentSvc *service.CommentModerationService
}

// NewReviewAdminHandler builds the handler over its two services.
func NewReviewAdminHandler(svc *service.ReviewModerationService, commentSvc *service.CommentModerationService) *ReviewAdminHandler {
	return &ReviewAdminHandler{svc: svc, commentSvc: commentSvc}
}

// RegisterRoutes attaches A1-A4 and A7 to the A0-gated admin group, plus a
// nested "/reviews/:id" group carrying A5/A6 — CommentAdminHandler owns
// those two handlers, this method only wires the route paths so the scoping
// review id is shared between both feature halves.
func (h *ReviewAdminHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/reviews", h.List)
	gated.GET("/reviews/:id", h.Get)
	gated.PATCH("/reviews/:id/status", h.UpdateStatus)
	gated.DELETE("/reviews/:id", h.Delete)
	gated.GET("/review-categories", h.ListCategories)

	nested := gated.Group("/reviews/:id")
	NewCommentAdminHandler(h.commentSvc).RegisterRoutes(nested)
}

func (h *ReviewAdminHandler) List(c echo.Context) error {
	p, err := parseReviewListParams(c)
	if err != nil {
		return err
	}
	page, err := h.svc.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, page)
}

func (h *ReviewAdminHandler) Get(c echo.Context) error {
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

// reviewStatusRequest is A3's explicit DTO — never bind straight into
// domain.Review (mass assignment).
type reviewStatusRequest struct {
	Status string `json:"status"`
}

func (h *ReviewAdminHandler) UpdateStatus(c echo.Context) error {
	id, actor, err := reviewActionContext(c)
	if err != nil {
		return err
	}
	var req reviewStatusRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	updated, err := h.svc.UpdateStatus(c.Request().Context(), actor, id, req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *ReviewAdminHandler) Delete(c echo.Context) error {
	id, actor, err := reviewActionContext(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), actor, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ReviewAdminHandler) ListCategories(c echo.Context) error {
	categories, err := h.svc.ListCategories(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, categories)
}

// reviewActionContext resolves A3/A4's two inputs: the target review id
// from the path (pathUUID — 400 on a malformed id) and the acting admin
// from the verified JWT subject, the same pattern as
// user_admin_handler.go's userActionContext.
func reviewActionContext(c echo.Context) (id, actor uuid.UUID, err error) {
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
