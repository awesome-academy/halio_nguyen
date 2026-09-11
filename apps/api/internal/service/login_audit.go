package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

func (s *AuthService) logActivityBestEffort(ctx context.Context, userID uuid.UUID, action, ip, userAgent string) {
	err := s.activity.Insert(ctx, s.db, repository.NewActivityLog{
		UserID:     userID,
		Action:     action,
		EntityType: domain.EntityTypeUser,
		EntityID:   userID,
		IPAddress:  ip,
		UserAgent:  userAgent,
	})
	if err != nil {
		slog.Warn("failed to write activity log", "action", action, "error", err)
	}
}

// logRejectedLogin never records a DB row (activity_logs.user_id is NOT
// NULL — a failed login has no subject to attach it to, decisions.md L1)
// and never logs the plaintext email, only a fixed-length fingerprint an
// operator can correlate across log lines without recovering the address.
func logRejectedLogin(email, ip string) {
	slog.Warn("admin login rejected", "email_hash", hashForLog(email), "ip", ip)
}

func hashForLog(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(email)))
	return hex.EncodeToString(sum[:8])
}
