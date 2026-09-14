package handler

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// pathParamUUID parses a named path param (e.g. "imageId", "scheduleId") as
// a UUID with a 400 before any query runs (Security: IDOR path confusion).
func pathParamUUID(c echo.Context, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest("Invalid " + name)
	}
	return id, nil
}

// parseTourListParams reads the shared list query plus F003's filters.
// Malformed numbers fall back to defaults (Normalize clamps them); a
// malformed category_id/price bound is a 400 since silently ignoring it
// would change which rows the admin sees.
func parseTourListParams(c echo.Context) (repository.TourListParams, error) {
	var p repository.TourListParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")
	p.Search = c.QueryParam("search")

	if raw := c.QueryParam("category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return p, apperror.NewBadRequest("category_id must be a valid id")
		}
		p.CategoryID = &id
	}
	if raw := c.QueryParam("status"); raw != "" {
		p.Status = &raw
	}
	if raw := c.QueryParam("price_min"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return p, apperror.NewBadRequest("price_min must be a number")
		}
		p.PriceMin = &v
	}
	if raw := c.QueryParam("price_max"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return p, apperror.NewBadRequest("price_max must be a number")
		}
		p.PriceMax = &v
	}
	return p, nil
}
