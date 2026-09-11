package config

import (
	"fmt"
	"os"
)

// minJWTSecretLen is the shortest secret echo-jwt is trusted to sign with —
// short of it, an HS256 signature is brute-forceable offline.
const minJWTSecretLen = 32

// devJWTSecretFallback is used only when APP_ENV == "development" and
// JWT_SECRET is unset, so a fresh local checkout can boot without config.
const devJWTSecretFallback = "super-secret-jwt-key-min-32-chars-default"

// resolveJWTSecret fails closed outside development: a missing or
// too-short JWT_SECRET aborts process start instead of silently signing
// admin sessions with a source-visible default (decisions.md/phase-02
// Security Considerations — "Secret strength — boot-time, fail-closed").
func resolveJWTSecret(appEnv string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	if (secret == "" || len(secret) < minJWTSecretLen) && appEnv != "development" {
		return "", fmt.Errorf(
			"JWT_SECRET must be set to at least %d characters outside development (APP_ENV=%s)",
			minJWTSecretLen, appEnv,
		)
	}

	if secret == "" {
		secret = devJWTSecretFallback
	}

	return secret, nil
}
