package jwtutil

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-at-least-32-bytes-long!"

func TestIssueAndParseRoundTrip(t *testing.T) {
	token, err := Issue(testSecret, time.Hour, "user-1", "admin")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	claims, err := Parse(testSecret, token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
}

func TestIssueExpiryMatchesTTL(t *testing.T) {
	token, err := Issue(testSecret, time.Hour, "user-1", "admin")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	claims, err := Parse(testSecret, token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	gotDelta := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if gotDelta != time.Hour {
		t.Errorf("exp - iat = %v, want %v (D2)", gotDelta, time.Hour)
	}
}

func TestIssueRejectsShortSecret(t *testing.T) {
	if _, err := Issue("too-short", time.Hour, "user-1", "admin"); err == nil {
		t.Fatal("expected Issue to reject a secret under 32 bytes")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	token, err := Issue(testSecret, -time.Minute, "user-1", "admin")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if _, err := Parse(testSecret, token); err == nil {
		t.Fatal("expected Parse to reject an expired token")
	}
}

func TestParseRejectsBadSignature(t *testing.T) {
	token, err := Issue(testSecret, time.Hour, "user-1", "admin")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if _, err := Parse("a-completely-different-secret-32b!", token); err == nil {
		t.Fatal("expected Parse to reject a token signed with a different secret")
	}
}

func TestParseRejectsAlgNone(t *testing.T) {
	claims := AdminClaims{
		UserID: "user-1",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build alg:none token: %v", err)
	}
	if !strings.Contains(tokenString, ".") {
		t.Fatal("malformed test token")
	}

	if _, err := Parse(testSecret, tokenString); err == nil {
		t.Fatal("expected Parse to reject an alg:none token")
	}
}
