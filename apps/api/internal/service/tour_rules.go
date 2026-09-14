package service

import (
	"strings"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// maxTourTitleLen/maxTourSlugLen mirror the DB column widths (tours.title
// VARCHAR(255), tours.slug VARCHAR(280)). maxTourSlugLen is wider than
// categories' 120 cap, but Slugify's own 120-char output is always safe here
// (slug.go's comment on why the same helper is reused verbatim).
const (
	maxTourTitleLen       = 255
	maxTourDestinationLen = 255
	maxHighlightCount     = 20
	maxHighlightLen       = 200
)

// buildTour validates a TourInput and shapes it into the row to write.
// BR-005 (avg_rating/total_ratings never admin-writable) is checked
// unconditionally: the Create handler DTO never binds those fields, so they
// are only ever non-nil when an Update request supplied them.
func buildTour(in TourInput) (*domain.Tour, error) {
	fields := map[string]string{}

	title := strings.TrimSpace(in.Title)
	switch {
	case title == "":
		fields["title"] = "Title is required."
	case len(title) > maxTourTitleLen:
		fields["title"] = "Title must be 255 characters or fewer."
	}

	// FR-002 / Security Considerations: the tour slug is always
	// server-derived from the title — unlike categories (FR-401), F003 gives
	// the admin no override, so TourInput carries no client-supplied slug.
	slug := Slugify(title)
	if slug == "" {
		fields["slug"] = "Slug is required."
	}

	destination := strings.TrimSpace(in.Destination)
	switch {
	case destination == "":
		fields["destination"] = "Destination is required."
	case len(destination) > maxTourDestinationLen:
		fields["destination"] = "Destination must be 255 characters or fewer."
	}

	if strings.TrimSpace(in.Description) == "" {
		fields["description"] = "Description is required."
	}

	// BR-002: duration_days and max_participants positive; duration_nights >= 0.
	if in.DurationDays <= 0 {
		fields["duration_days"] = "Duration (days) must be greater than zero."
	}
	if in.DurationNights < 0 {
		fields["duration_nights"] = "Duration (nights) must be zero or more."
	}
	if in.MaxParticipants <= 0 {
		fields["max_participants"] = "Max participants must be greater than zero."
	}

	// Mirrors the DB CHECK (price >= 0); pre-validated for a clean 422
	// instead of a raw constraint error.
	if in.Price < 0 {
		fields["price"] = "Price must be zero or more."
	}
	// BR-001: discount_price, when set, must be <= price (and >= 0).
	if in.DiscountPrice != nil {
		switch {
		case *in.DiscountPrice < 0:
			fields["discount_price"] = "Discounted price must be zero or more."
		case *in.DiscountPrice > in.Price:
			fields["discount_price"] = "Discounted price cannot exceed the regular price."
		}
	}

	if len(in.Highlights) > maxHighlightCount {
		fields["highlights"] = "No more than 20 highlights are allowed."
	}
	for _, h := range in.Highlights {
		if len(h) > maxHighlightLen {
			fields["highlights"] = "Each highlight must be 200 characters or fewer."
			break
		}
	}

	if in.ThumbnailURL != nil && *in.ThumbnailURL != "" && !isHTTPURL(*in.ThumbnailURL) {
		fields["thumbnail_url"] = "Please enter a valid http(s) URL."
	}

	// BR-005: avg_rating/total_ratings are never admin-writable.
	if in.AvgRating != nil {
		fields["avg_rating"] = "Average rating is read-only."
	}
	if in.TotalRatings != nil {
		fields["total_ratings"] = "Total ratings is read-only."
	}

	if len(fields) > 0 {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", fields)
	}

	status := domain.TourStatusDraft
	if in.Status != nil && *in.Status != "" {
		status = *in.Status
	}
	if !validTourStatus(status) {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", map[string]string{"status": "Invalid status."})
	}

	return &domain.Tour{
		CategoryID:      in.CategoryID,
		Title:           title,
		Slug:            slug,
		Description:     strings.TrimSpace(in.Description),
		Itinerary:       emptyToNil(in.Itinerary),
		Destination:     destination,
		DurationDays:    in.DurationDays,
		DurationNights:  in.DurationNights,
		Price:           in.Price,
		DiscountPrice:   in.DiscountPrice,
		MaxParticipants: in.MaxParticipants,
		ThumbnailURL:    emptyToNil(in.ThumbnailURL),
		Highlights:      in.Highlights,
		Inclusions:      emptyToNil(in.Inclusions),
		Exclusions:      emptyToNil(in.Exclusions),
		Status:          status,
	}, nil
}

func validTourStatus(s string) bool {
	switch s {
	case domain.TourStatusDraft, domain.TourStatusPublished, domain.TourStatusArchived:
		return true
	default:
		return false
	}
}

// tourTransitions is SM-001: draft->published, published->archived,
// archived->published (F003 §5 accepted default: reactivation is allowed).
var tourTransitions = map[string][]string{
	domain.TourStatusDraft:     {domain.TourStatusPublished},
	domain.TourStatusPublished: {domain.TourStatusArchived},
	domain.TourStatusArchived:  {domain.TourStatusPublished},
}

func canTransition(table map[string][]string, from, to string) bool {
	for _, allowed := range table[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
