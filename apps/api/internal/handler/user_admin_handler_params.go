package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// parseUserListParams reads the shared list query plus F005's role/
// is_active filters. Malformed numbers fall back to defaults (Normalize
// clamps them); a malformed is_active is a 400, since silently ignoring it
// would change which accounts the admin sees. role is left unvalidated
// here — an out-of-enum value simply matches zero rows on this read path,
// while A3's write path validates it server-side (422) before it ever
// reaches SQL.
func parseUserListParams(c echo.Context) (repository.UserListParams, error) {
	var p repository.UserListParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")
	p.Search = c.QueryParam("search")

	if raw := c.QueryParam("role"); raw != "" {
		p.Role = &raw
	}
	if raw := c.QueryParam("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return p, apperror.NewBadRequest("is_active must be true or false")
		}
		p.IsActive = &v
	}
	return p, nil
}
