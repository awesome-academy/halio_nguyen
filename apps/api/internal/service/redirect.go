package service

import "strings"

// defaultRedirect is where a login lands when no from path was supplied or
// the supplied one fails ValidateRedirect (DEC-001).
const defaultRedirect = "/admin/dashboard"

// ValidateRedirect returns from unchanged only when it is a same-origin
// admin path with no traversal segment; otherwise it falls back to
// defaultRedirect (DEC-001). Rejects //host, /\host, an explicit scheme,
// anything outside /admin/*, and any "/../" segment even after the prefix
// checks pass — /admin/../../x satisfies every other rule and must be
// caught by the segment scan (Security Considerations, decisions.md #9
// finding 14).
func ValidateRedirect(from string) string {
	if from == "" {
		return defaultRedirect
	}
	// Browsers normalize `\` to `/` when parsing a same-origin URL, so
	// `/admin/..\..\x` would pass a `/`-only segment scan and still escape
	// /admin/ once router.replace runs it. A legitimate admin path never
	// contains a backslash, so reject any occurrence outright.
	if strings.Contains(from, `\`) {
		return defaultRedirect
	}
	if !strings.HasPrefix(from, "/") || strings.HasPrefix(from, "//") {
		return defaultRedirect
	}
	if strings.Contains(from, "://") {
		return defaultRedirect
	}
	if !strings.HasPrefix(from, "/admin/") {
		return defaultRedirect
	}
	for _, segment := range strings.Split(from, "/") {
		if segment == ".." {
			return defaultRedirect
		}
	}
	return from
}
