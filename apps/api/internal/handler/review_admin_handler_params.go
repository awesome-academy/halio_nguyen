package handler

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// parseReviewListParams reads the shared list query plus F006's status/
// review_category_id filters (A1). Malformed numbers fall back to defaults
// (Normalize clamps them); a malformed review_category_id is a 400 since
// silently ignoring it would change which reviews the admin sees. status is
// left unvalidated here — an out-of-enum value simply matches zero rows on
// this read path, while A3's write path validates it server-side (422)
// before it ever reaches SQL.
func parseReviewListParams(c echo.Context) (repository.ReviewListParams, error) {
	var p repository.ReviewListParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")
	p.Search = c.QueryParam("search")

	if raw := c.QueryParam("status"); raw != "" {
		p.Status = &raw
	}
	if raw := c.QueryParam("review_category_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return p, apperror.NewBadRequest("review_category_id must be a valid UUID")
		}
		p.ReviewCategoryID = &id
	}
	return p, nil
}
