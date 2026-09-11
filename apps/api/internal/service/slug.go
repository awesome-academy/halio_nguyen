package service

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// maxSlugLen matches categories.slug VARCHAR(120) (tours.slug is wider, so
// the same cap is safe for Phase 4's tour slugs).
const maxSlugLen = 120

// Slugify derives a URL slug in Go rather than via Postgres' f_unaccent()
// so the service stays testable without a database (D-A7): NFD-normalise,
// drop combining marks, map đ/Đ (which NFD does not decompose), lowercase,
// collapse every non-alphanumeric run into one "-", trim, cap length.
func Slugify(s string) string {
	s = strings.NewReplacer("đ", "d", "Đ", "D").Replace(s)

	var b strings.Builder
	pendingDash := false
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		r = unicode.ToLower(r)
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if !isAlnum {
			pendingDash = b.Len() > 0
			continue
		}
		if pendingDash {
			b.WriteByte('-')
			pendingDash = false
		}
		b.WriteRune(r)
	}

	out := b.String()
	if len(out) > maxSlugLen {
		out = strings.TrimRight(out[:maxSlugLen], "-")
	}
	return out
}
