package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
)

func TestCategorySortAllowlistNeverEchoesInput(t *testing.T) {
	col, dir := ResolveSort(categorySortAllow, "name;DROP TABLE categories", "desc")
	if col != "c.sort_order" {
		t.Errorf("hostile sort_by resolved to %q, want the default c.sort_order", col)
	}
	if dir != SortDesc {
		t.Errorf("dir = %q, want desc", dir)
	}

	col, dir = ResolveSort(categorySortAllow, "created_at", "")
	if col != "c.created_at" || dir != SortAsc {
		t.Errorf("created_at resolved to %q %q", col, dir)
	}
}

func TestUniqueViolationFieldMapsKnownConstraints(t *testing.T) {
	err := &pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "categories_slug_key"}
	field, ok := UniqueViolationField(err, categoryUniqueConstraints)
	if !ok || field != "slug" {
		t.Fatalf("got (%q, %v), want (slug, true)", field, ok)
	}

	if _, ok := UniqueViolationField(&pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "other_key"}, categoryUniqueConstraints); ok {
		t.Error("an unmapped constraint must not be reported as a field conflict")
	}
	if _, ok := UniqueViolationField(errors.New("boom"), categoryUniqueConstraints); ok {
		t.Error("a non-pg error must not be reported as a field conflict")
	}
}

func TestCategoryConflictIsFieldScoped409(t *testing.T) {
	err := CategoryConflict("name")
	if err.Code != apperror.CodeConflict {
		t.Errorf("code = %v, want conflict", err.Code)
	}
	if err.Fields["name"] == "" {
		t.Errorf("expected fields.name populated, got %v", err.Fields)
	}
	if _, leaked := err.Fields["categories_name_key"]; leaked {
		t.Error("constraint name must never reach the body")
	}
}
