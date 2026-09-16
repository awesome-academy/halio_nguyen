package handler

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// bookingDateLayout is the wire format for date_from/date_to. Both bound
// bookings.created_at (FR-202's created-date range), which is TIMESTAMPTZ —
// date_to is widened to the end of its day below so an admin filtering
// "to today" still sees bookings made earlier today.
const bookingDateLayout = "2006-01-02"

// parseBookingListParams reads the shared list query plus F004's filters.
// Malformed numbers fall back to defaults (Normalize clamps them); a
// malformed id or date is rejected, since silently ignoring it would change
// which bookings the admin sees.
func parseBookingListParams(c echo.Context) (repository.BookingListParams, error) {
	var p repository.BookingListParams
	p.Page, _ = strconv.Atoi(c.QueryParam("page"))
	p.PageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	p.SortBy = c.QueryParam("sort_by")
	p.SortDir = c.QueryParam("sort_dir")
	p.Search = c.QueryParam("search")

	if raw := c.QueryParam("status"); raw != "" {
		p.Status = &raw
	}
	if raw := c.QueryParam("tour_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return p, apperror.NewBadRequest("tour_id must be a valid id")
		}
		p.TourID = &id
	}
	if raw := c.QueryParam("schedule_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return p, apperror.NewBadRequest("schedule_id must be a valid id")
		}
		p.ScheduleID = &id
	}

	from, err := parseBookingDate(c.QueryParam("date_from"), false)
	if err != nil {
		return p, apperror.NewBadRequest("date_from must be YYYY-MM-DD")
	}
	to, err := parseBookingDate(c.QueryParam("date_to"), true)
	if err != nil {
		return p, apperror.NewBadRequest("date_to must be YYYY-MM-DD")
	}

	// An inverted range is rejected before the query runs rather than
	// silently returning an empty page, which reads as "no bookings" when
	// the truth is "your filter is backwards".
	if from != nil && to != nil && to.Before(*from) {
		msg := "The start date must be on or before the end date."
		return p, apperror.NewUnprocessable(msg, map[string]string{"date_from": msg})
	}
	p.DateFrom, p.DateTo = from, to

	return p, nil
}

// parseBookingDate parses a YYYY-MM-DD bound. endOfDay pushes the value to
// 23:59:59.999999999 so date_to is inclusive of the whole day against a
// TIMESTAMPTZ column.
func parseBookingDate(raw string, endOfDay bool) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(bookingDateLayout, raw)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.AddDate(0, 0, 1).Add(-time.Nanosecond)
	}
	return &t, nil
}
