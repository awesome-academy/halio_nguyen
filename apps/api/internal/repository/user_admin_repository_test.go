package repository

import "testing"

// SQL-shape/allowlist table test in the style of category_repository_test.go
// — a pure function test against userSortAllow, since CI has no Postgres
// (L5) to exercise the real query against.
func TestUserSortAllowlistNeverEchoesInput(t *testing.T) {
	col, dir := ResolveSort(userSortAllow, "role;DROP TABLE users", "desc")
	if col != "created_at" {
		t.Errorf("hostile sort_by resolved to %q, want the default created_at", col)
	}
	if dir != SortDesc {
		t.Errorf("dir = %q, want desc", dir)
	}

	col, dir = ResolveSort(userSortAllow, "email", "")
	if col != "email" || dir != SortAsc {
		t.Errorf("email resolved to %q %q", col, dir)
	}

	for _, name := range []string{"full_name", "role", "is_active", "created_at"} {
		if _, ok := userSortAllow[name]; !ok {
			t.Errorf("allowlist missing %q", name)
		}
	}
}
