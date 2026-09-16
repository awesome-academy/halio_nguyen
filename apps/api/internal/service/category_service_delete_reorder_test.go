package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Delete (BR-001 / D4) and reorder (ALG-001) coverage. Split out of
// category_service_test.go, which holds the create/update cases and the
// shared fakes both files use.

func TestDeleteBlockedByLiveToursWritesNothing(t *testing.T) {
	repo := &mockCategoryRepo{tourCount: 2}
	svc, tx := newSvc(repo)

	err := svc.Delete(context.Background(), uuid.New())
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeConflict {
		t.Fatalf("want 409, got %v", err)
	}
	if repo.softDeletes != 0 || tx.committed || !tx.rolledBack {
		t.Errorf("blocked delete must write nothing: softDeletes=%d committed=%v rolledBack=%v", repo.softDeletes, tx.committed, tx.rolledBack)
	}
}

func TestDeleteSucceedsWithZeroTours(t *testing.T) {
	repo := &mockCategoryRepo{tourCount: 0}
	svc, tx := newSvc(repo)
	if err := svc.Delete(context.Background(), uuid.New()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if repo.softDeletes != 1 || !tx.committed {
		t.Errorf("softDeletes=%d committed=%v", repo.softDeletes, tx.committed)
	}
}

func TestDeleteUnknownCategoryIs404(t *testing.T) {
	svc, _ := newSvc(&mockCategoryRepo{lockErr: repository.ErrCategoryNotFound})
	err := svc.Delete(context.Background(), uuid.New())
	if appErr, ok := err.(*apperror.Error); !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("want 404, got %v", err)
	}
}

func seeded() []domain.Category {
	// Seed data is 1-based with the ids a..c in display order.
	return []domain.Category{
		{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000a"), SortOrder: 1},
		{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000b"), SortOrder: 2},
		{ID: uuid.MustParse("00000000-0000-0000-0000-00000000000c"), SortOrder: 3},
	}
}

func assertRanks(t *testing.T, got []domain.Category, wantIDs ...string) {
	t.Helper()
	if len(got) != len(wantIDs) {
		t.Fatalf("len = %d, want %d", len(got), len(wantIDs))
	}
	for i, c := range got {
		if c.SortOrder != i {
			t.Errorf("rank %d: sort_order = %d (must be contiguous 0..n-1)", i, c.SortOrder)
		}
		if c.ID.String()[35:] != wantIDs[i] {
			t.Errorf("rank %d: id ends %q, want %q", i, c.ID.String()[35:], wantIDs[i])
		}
	}
}

func TestReorderMovesUpAndRewritesRanks(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, tx := newSvc(repo)
	got, err := svc.Reorder(context.Background(), seeded()[2].ID, 0)
	if err != nil {
		t.Fatalf("Reorder: %v", err)
	}
	assertRanks(t, got, "c", "a", "b")
	if !tx.committed {
		t.Error("expected commit")
	}
}

func TestReorderClampsBothBoundaries(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, _ := newSvc(repo)
	got, _ := svc.Reorder(context.Background(), seeded()[0].ID, 99)
	assertRanks(t, got, "b", "c", "a")

	repo = &mockCategoryRepo{reorderSet: seeded()}
	svc, _ = newSvc(repo)
	got, _ = svc.Reorder(context.Background(), seeded()[1].ID, -5)
	assertRanks(t, got, "b", "a", "c")
}

func TestReorderMoveDownAndUnknownID(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, _ := newSvc(repo)
	got, _ := svc.Reorder(context.Background(), seeded()[0].ID, 1)
	assertRanks(t, got, "b", "a", "c")

	_, err := svc.Reorder(context.Background(), uuid.New(), 0)
	if appErr, ok := err.(*apperror.Error); !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("unknown id: want 404, got %v", err)
	}
}
