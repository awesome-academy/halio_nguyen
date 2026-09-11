package service

import "testing"

func TestValidateRedirectHostileInputs(t *testing.T) {
	cases := map[string]string{
		"":                     defaultRedirect,
		"//evil.tld":           defaultRedirect,
		`/\evil.tld`:           defaultRedirect,
		"https://evil.tld":     defaultRedirect,
		"/admin/../../x":       defaultRedirect,
		`/admin/..\..\evil`:    defaultRedirect,
		`/admin\evil`:          defaultRedirect,
		// Accepted on purpose: the server never URL-decodes `from`, and the
		// browser does not treat a literal %2f as a path separator, so this
		// stays inside /admin/ — a documented non-bypass, not a gap.
		"/admin/..%2f..%2fx": "/admin/..%2f..%2fx",
		"/not-admin/dashboard": defaultRedirect,
		"/admin/tours":         "/admin/tours",
	}
	for input, want := range cases {
		if got := ValidateRedirect(input); got != want {
			t.Errorf("ValidateRedirect(%q) = %q, want %q", input, got, want)
		}
	}
}
