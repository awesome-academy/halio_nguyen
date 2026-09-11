package service

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

const (
	maxCategoryNameLen = 100
	maxCategorySlugLen = 120
)

var slugShape = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// buildCategory validates a CategoryInput and shapes it into the row to
// write. Slug is derived from name only when the client sent none (FR-401
// lets the admin override). Semantic failures are 422 with field errors.
func buildCategory(in CategoryInput) (*domain.Category, error) {
	fields := map[string]string{}

	name := strings.TrimSpace(in.Name)
	switch {
	case name == "":
		fields["name"] = "Name is required."
	case len(name) > maxCategoryNameLen:
		fields["name"] = "Name must be 100 characters or fewer."
	}

	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = Slugify(name)
	}
	switch {
	case slug == "":
		fields["slug"] = "Slug is required."
	case len(slug) > maxCategorySlugLen:
		fields["slug"] = "Slug must be 120 characters or fewer."
	case !slugShape.MatchString(slug):
		fields["slug"] = "Slug must be lowercase letters, numbers, and hyphens only."
	}

	if in.ImageURL != nil && *in.ImageURL != "" && !isHTTPURL(*in.ImageURL) {
		fields["image_url"] = "Please enter a valid http(s) URL."
	}

	if len(fields) > 0 {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", fields)
	}

	c := &domain.Category{
		Name:        name,
		Slug:        slug,
		Description: emptyToNil(in.Description),
		ImageURL:    emptyToNil(in.ImageURL),
		IsActive:    true,
	}
	if in.SortOrder != nil {
		c.SortOrder = *in.SortOrder
	}
	if in.IsActive != nil {
		c.IsActive = *in.IsActive
	}
	return c, nil
}

// isHTTPURL rejects javascript:/data:/relative values: image_url is
// rendered into <img src>, so the server — not only the form — must gate it.
func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func emptyToNil(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

func indexOfCategory(cats []domain.Category, id uuid.UUID) int {
	for i := range cats {
		if cats[i].ID == id {
			return i
		}
	}
	return -1
}

// clampRank bounds an ALG-001 target to [0, n-1]; out-of-range input is
// clamped, never a 500 (technical-spec § 3.2).
func clampRank(target, n int) int {
	if target < 0 {
		return 0
	}
	if target > n-1 {
		return n - 1
	}
	return target
}

// moveCategory returns a new slice with element cur relocated to target.
func moveCategory(cats []domain.Category, cur, target int) []domain.Category {
	out := make([]domain.Category, 0, len(cats))
	moved := cats[cur]
	for i := range cats {
		if i != cur {
			out = append(out, cats[i])
		}
	}
	out = append(out[:target], append([]domain.Category{moved}, out[target:]...)...)
	return out
}
