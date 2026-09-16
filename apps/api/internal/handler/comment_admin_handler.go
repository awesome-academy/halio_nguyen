package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	adminmw "github.com/sun-booking/sun-booking-tours/apps/api/internal/middleware"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// CommentAdminHandler is F006's A5/A6: hide/unhide and soft-delete one
// comment. It is registered by ReviewAdminHandler.RegisterRoutes on a
// nested group that already carries the "/reviews/:id" prefix, so
// router.go still gains only one RegisterRoutes call for the whole feature.
type CommentAdminHandler struct {
	svc *service.CommentModerationService
}

// NewCommentAdminHandler builds the handler over its service.
func NewCommentAdminHandler(svc *service.CommentModerationService) *CommentAdminHandler {
	return &CommentAdminHandler{svc: svc}
}

// RegisterRoutes attaches A5/A6 under nested, which already carries
// "/reviews/:id" — the full paths become
// "/reviews/:id/comments/:commentId/visibility" and
// "/reviews/:id/comments/:commentId".
func (h *CommentAdminHandler) RegisterRoutes(nested *echo.Group) {
	nested.PATCH("/comments/:commentId/visibility", h.UpdateVisibility)
	nested.DELETE("/comments/:commentId", h.Delete)
}

// commentVisibilityRequest is A5's explicit DTO — never bind straight into
// domain.Comment (mass assignment). IsHidden is a pointer so a missing
// field is distinguishable from an explicit false.
type commentVisibilityRequest struct {
	IsHidden *bool `json:"is_hidden"`
}

func (h *CommentAdminHandler) UpdateVisibility(c echo.Context) error {
	reviewID, commentID, actor, err := commentActionContext(c)
	if err != nil {
		return err
	}
	var req commentVisibilityRequest
	if err := c.Bind(&req); err != nil || req.IsHidden == nil {
		return apperror.NewBadRequest("is_hidden is required")
	}
	updated, err := h.svc.UpdateVisibility(c.Request().Context(), actor, reviewID, commentID, *req.IsHidden)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *CommentAdminHandler) Delete(c echo.Context) error {
	reviewID, commentID, actor, err := commentActionContext(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), actor, reviewID, commentID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// commentActionContext resolves A5/A6's three inputs: the review id and
// comment id from the path (either malformed -> 400) and the acting admin
// from the verified JWT subject, never a caller-supplied value. The
// review/comment pairing itself is enforced downstream by
// CommentAdminRepository's `AND review_id = @review_id` predicate — a
// mismatched pair reaches SQL and comes back 404, rather than being
// rejected here (R6/IDOR).
func commentActionContext(c echo.Context) (reviewID, commentID, actor uuid.UUID, err error) {
	reviewID, err = uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperror.NewBadRequest("Invalid review id")
	}
	commentID, err = uuid.Parse(c.Param("commentId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperror.NewBadRequest("Invalid comment id")
	}
	claims, ok := adminmw.ClaimsFrom(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	actor, err = uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperror.NewUnauthorized("Your session has expired or is invalid. Please log in again.")
	}
	return reviewID, commentID, actor, nil
}
