package repository

import (
	"errors"
	"testing"
)

// CI has no Postgres (L5), so the guarded UPDATE statements themselves
// cannot be exercised here — that behaviour is proven by
// comment_moderation_service_test.go against a mock, and by the manual curl
// pass (Implementation Steps #7). This file asserts the one thing that is
// pure Go: ErrCommentNotFound is its own sentinel, distinct from every other
// repository's not-found error, so a caller's errors.Is check can never
// cross-match the wrong resource.
func TestErrCommentNotFoundIsADistinctSentinel(t *testing.T) {
	if ErrCommentNotFound == nil {
		t.Fatal("ErrCommentNotFound must not be nil")
	}
	if ErrCommentNotFound.Error() == "" {
		t.Fatal("ErrCommentNotFound must carry a message")
	}
	if errors.Is(ErrCommentNotFound, ErrReviewNotFound) {
		t.Error("ErrCommentNotFound must not equal ErrReviewNotFound")
	}
	if errors.Is(ErrCommentNotFound, ErrUserNotFound) {
		t.Error("ErrCommentNotFound must not equal ErrUserNotFound")
	}
}
