package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func extractWith(t *testing.T, cidrs []string, remoteAddr, xff string) string {
	t.Helper()
	extractor, err := NewIPExtractor(cidrs)
	if err != nil {
		t.Fatalf("NewIPExtractor: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	return extractor(req)
}

func TestIPExtractorTrustsXFFFromListedProxy(t *testing.T) {
	got := extractWith(t, []string{"127.0.0.1/32"}, "127.0.0.1:4000", "203.0.113.9")
	if got != "203.0.113.9" {
		t.Errorf("RealIP = %q, want the XFF client 203.0.113.9", got)
	}
}

func TestIPExtractorIgnoresXFFFromUnlistedPrivateHost(t *testing.T) {
	// 10.0.0.5 is RFC1918 — echo would trust it by default; we must not.
	got := extractWith(t, []string{"127.0.0.1/32"}, "10.0.0.5:4000", "203.0.113.9")
	if got != "10.0.0.5" {
		t.Errorf("RealIP = %q, want the untrusted peer 10.0.0.5 (XFF must be ignored)", got)
	}
}

func TestIPExtractorRejectsMalformedCIDR(t *testing.T) {
	if _, err := NewIPExtractor([]string{"not-a-cidr"}); err == nil {
		t.Fatal("expected a malformed CIDR to fail closed at boot")
	}
}
