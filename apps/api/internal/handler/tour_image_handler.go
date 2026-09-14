package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// TourImageHandler is the HTTP entry point for F003's gallery actions (A7-A9).
type TourImageHandler struct {
	svc *service.TourImageService
}

// NewTourImageHandler builds the handler over its service.
func NewTourImageHandler(svc *service.TourImageService) *TourImageHandler {
	return &TourImageHandler{svc: svc}
}

// RegisterRoutes attaches A7-A9 to the A0-gated admin group.
func (h *TourImageHandler) RegisterRoutes(gated *echo.Group) {
	gated.POST("/tours/:id/images", h.Create)
	gated.PUT("/tours/:id/images/:imageId", h.Update)
	gated.DELETE("/tours/:id/images/:imageId", h.Delete)
}

// tourImageRequest is A7's create DTO; also reused by TourHandler for A3's
// nested images[] array (same shape, same package).
type tourImageRequest struct {
	ImageURL string  `json:"image_url"`
	Caption  *string `json:"caption"`
}

func (r tourImageRequest) toInput() service.TourImageInput {
	return service.TourImageInput{ImageURL: r.ImageURL, Caption: r.Caption}
}

// tourImageUpdateRequest is A8's DTO (BR-011: sort_order is required and
// trusted verbatim — never silently defaulted).
type tourImageUpdateRequest struct {
	Caption   *string `json:"caption"`
	SortOrder *int    `json:"sort_order"`
}

func (h *TourImageHandler) Create(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req tourImageRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	created, err := h.svc.Create(c.Request().Context(), tourID, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *TourImageHandler) Update(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	imageID, err := pathParamUUID(c, "imageId")
	if err != nil {
		return err
	}
	var req tourImageUpdateRequest
	if err := c.Bind(&req); err != nil || req.SortOrder == nil {
		return apperror.NewBadRequest("sort_order is required")
	}
	updated, err := h.svc.Update(c.Request().Context(), tourID, imageID, service.TourImageUpdateInput{
		Caption: req.Caption, SortOrder: *req.SortOrder,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *TourImageHandler) Delete(c echo.Context) error {
	tourID, err := pathUUID(c)
	if err != nil {
		return err
	}
	imageID, err := pathParamUUID(c, "imageId")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), tourID, imageID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
