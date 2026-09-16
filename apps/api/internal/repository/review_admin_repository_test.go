package repository

import (
	"errors"
	"testing"
)

// SQL-shape/allowlist table test in the style of
// user_admin_repository_test.go — CI has no Postgres (L5) to exercise the
// real query against, so ResolveSort's allowlist behaviour is what proves
// BR-005's ORDER BY injection defence.
func TestReviewSortAllowlistNeverEchoesInput(t *testing.T) {
	col, dir := ResolveSort(reviewSortAllow, "title;DROP TABLE reviews", "desc")
	if col != "created_at" {
		t.Errorf("hostile sort_by resolved to %q, want the default created_at", col)
	}
	if dir != SortDesc {
		t.Errorf("dir = %q, want desc", dir)
	}

	col, dir = ResolveSort(reviewSortAllow, "like_count", "")
	if col != "like_count" || dir != SortAsc {
		t.Errorf("like_count resolved to %q %q", col, dir)
	}

	for _, name := range []string{"created_at", "like_count", "comment_count", "title"} {
		if _, ok := reviewSortAllow[name]; !ok {
			t.Errorf("allowlist missing %q", name)
		}
	}
}

func TestErrReviewNotFoundIsADistinctSentinel(t *testing.T) {
	if ErrReviewNotFound == nil {
		t.Fatal("ErrReviewNotFound must not be nil")
	}
	if ErrReviewNotFound.Error() == "" {
		t.Fatal("ErrReviewNotFound must carry a message")
	}
	if errors.Is(ErrReviewNotFound, ErrUserNotFound) {
		t.Error("ErrReviewNotFound must not equal ErrUserNotFound")
	}
}
