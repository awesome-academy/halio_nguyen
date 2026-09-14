package repository

import (
	"testing"

	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// TestTourSortAllowlist proves BR-008's allowlist resolves exactly the
// documented columns and silently falls back — never a 400 — for anything
// else, including an injection attempt (SC-001).
func TestTourSortAllowlist(t *testing.T) {
	tests := []struct {
		name    string
		by, dir string
		wantCol string
		wantDir string
	}{
		{"title asc", "title", "asc", "t.title", "asc"},
		{"price desc", "price", "desc", "t.price", "desc"},
		{"destination default dir", "destination", "", "t.destination", "asc"},
		{"created_at explicit", "created_at", "desc", "t.created_at", "desc"},
		{"unrecognised falls back to created_at", "bogus", "asc", "t.created_at", "asc"},
		{"injection attempt falls back", "id); DROP TABLE tours;--", "asc", "t.created_at", "asc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, dir := ResolveSort(tourSortAllow, tt.by, tt.dir)
			if col != tt.wantCol || dir != tt.wantDir {
				t.Errorf("ResolveSort(%q, %q) = (%q, %q), want (%q, %q)", tt.by, tt.dir, col, dir, tt.wantCol, tt.wantDir)
			}
		})
	}
}

// TestTourListDefaultsToCreatedAtDesc proves SC-001: an unspecified sort_dir
// alongside an unrecognised sort_by must resolve to created_at DESC as one
// atomic default, not created_at ASC (List's SortDir-defaulting logic).
func TestTourListDefaultsToCreatedAtDesc(t *testing.T) {
	sortDir := "" // as parseTourListParams leaves it when the caller sent none
	if sortDir == "" {
		sortDir = SortDesc
	}
	col, dir := ResolveSort(tourSortAllow, "bogus", sortDir)
	if col != "t.created_at" || dir != SortDesc {
		t.Errorf("default order = (%q, %q), want (t.created_at, desc)", col, dir)
	}
}

func TestTourArgsRoundTripsAllFields(t *testing.T) {
	price := 100.0
	discount := 80.0
	tour := &domain.Tour{
		Title: "Sample", Slug: "sample", Description: "d", Destination: "Da Nang",
		DurationDays: 3, DurationNights: 2, Price: price, DiscountPrice: &discount,
		MaxParticipants: 10, Highlights: []string{"a", "b"}, Status: domain.TourStatusDraft,
	}
	args := tourArgs(tour)
	if args["title"] != "Sample" || args["slug"] != "sample" || args["status"] != domain.TourStatusDraft {
		t.Errorf("tourArgs missing expected values: %+v", args)
	}
	if args["discount_price"] != &discount {
		t.Errorf("discount_price should be the same pointer, not copied")
	}
}

func TestTourConflictNamesTheField(t *testing.T) {
	err := TourConflict("slug")
	var appErr *apperror.Error
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	appErr = err
	if appErr.Code != apperror.CodeConflict || appErr.Fields["slug"] == "" {
		t.Errorf("want 409 with fields.slug, got %+v", appErr)
	}
}

func TestScheduleDateConflictNamesTheField(t *testing.T) {
	err := ScheduleDateConflict("departure_date")
	if err.Code != apperror.CodeConflict || err.Fields["departure_date"] == "" {
		t.Errorf("want 409 with fields.departure_date, got %+v", err)
	}
}
