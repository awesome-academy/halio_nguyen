package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// CategoryHandler is the HTTP entry point for F002 (A1–A5).
type CategoryHandler struct {
	svc *service.CategoryService
}

// NewCategoryHandler builds the handler over its service.
func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// RegisterRoutes attaches all five routes to the A0-gated admin group.
func (h *CategoryHandler) RegisterRoutes(gated *echo.Group) {
	gated.GET("/categories", h.List)
	gated.POST("/categories", h.Create)
	gated.PUT("/categories/:id", h.Update)
	gated.PATCH("/categories/:id/sort-order", h.Reorder)
	gated.DELETE("/categories/:id", h.Delete)
}

// categoryRequest is the explicit DTO — never domain.Category (mass
// assignment: id/timestamps are not bindable).
type categoryRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
	SortOrder   *int    `json:"sort_order"`
	IsActive    *bool   `json:"is_active"`
}

func (r categoryRequest) toInput() service.CategoryInput {
	return service.CategoryInput{
		Name: r.Name, Slug: r.Slug, Description: r.Description, ImageURL: r.ImageURL,
		SortOrder: r.SortOrder, IsActive: r.IsActive,
	}
}

type reorderRequest struct {
	SortOrder *int `json:"sort_order"`
}

func (h *CategoryHandler) List(c echo.Context) error {
	p, err := parseCategoryListParams(c)
	if err != nil {
		return err
	}
	page, err := h.svc.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, page)
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req categoryRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	created, err := h.svc.Create(c.Request().Context(), req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, created)
}

func (h *CategoryHandler) Update(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req categoryRequest
	if err := c.Bind(&req); err != nil {
		return apperror.NewBadRequest("Invalid request body")
	}
	updated, err := h.svc.Update(c.Request().Context(), id, req.toInput())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updated)
}

func (h *CategoryHandler) Reorder(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	var req reorderRequest
	if err := c.Bind(&req); err != nil || req.SortOrder == nil {
		return apperror.NewBadRequest("sort_order is required")
	}
	items, err := h.svc.Reorder(c.Request().Context(), id, *req.SortOrder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{"items": items})
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id, err := pathUUID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func pathUUID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest("Invalid id")
	}
	return id, nil
}

// parseCategoryListParams reads the shared list query plus is_active.
// Malformed numbers fall back to defaults (Normalize clamps them); only a
// malformed is_active is a 400, since silently ignoring it would change
// which rows the admin sees.
func parseCategoryListParams(c echo.Context) (repository.CategoryListParams, error) {
	var p repository.CategoryListParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")
	p.Search = c.QueryParam("search")

	if raw := c.QueryParam("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return p, apperror.NewBadRequest("is_active must be true or false")
		}
		p.IsActive = &v
	}
	return p, nil
}
