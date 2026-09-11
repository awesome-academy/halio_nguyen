// Package jwtutil issues and parses the admin session JWT. It is kept out of
// internal/middleware so both internal/service (issuing) and
// internal/middleware (verifying via echo-jwt) can share it without either
// side importing echo.
package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// minSecretLen mirrors config.minJWTSecretLen; duplicated here (not
// imported) because config already validates it at boot — this is a
// second, cheap defense against a future caller that builds a JWTConfig by
// hand instead of through config.Load().
const minSecretLen = 32

// AdminClaims is the JWT payload for an admin session. Role is checked by
// middleware.RequireAdminRole on every gated request.
type AdminClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Issue signs a new HS256 admin session token for userID/role, valid for
// ttl from now. It refuses to sign with a secret under minSecretLen so a
// misconfigured caller fails loudly instead of shipping a brute-forceable
// token.
func Issue(secret string, ttl time.Duration, userID, role string) (string, error) {
	if len(secret) < minSecretLen {
		return "", fmt.Errorf("jwtutil: secret must be at least %d bytes", minSecretLen)
	}

	now := time.Now()
	claims := AdminClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Parse verifies and decodes a signed admin token, pinning HS256 so an
// alg:none or RS256-confusion token cannot pass (see middleware.NewJWTConfig
// for the equivalent guard on the request-verification path).
func Parse(secret, tokenString string) (*AdminClaims, error) {
	claims := &AdminClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("jwtutil: token is invalid")
	}

	return claims, nil
}
