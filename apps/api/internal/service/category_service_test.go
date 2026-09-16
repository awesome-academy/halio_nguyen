package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// fakeTx/fakeDB stand in for the pool so transactional paths run without
// Postgres (D-A7). Embedding the interfaces means any method the mocked
// repository does not intercept would panic loudly rather than pass silently.
type fakeTx struct {
	pgx.Tx
	committed, rolledBack bool
}

func (t *fakeTx) Commit(context.Context) error   { t.committed = true; return nil }
func (t *fakeTx) Rollback(context.Context) error { t.rolledBack = true; return nil }

type fakeDB struct {
	repository.DB
	tx *fakeTx
}

func (d *fakeDB) Begin(context.Context) (pgx.Tx, error) { return d.tx, nil }

type mockCategoryRepo struct {
	repository.CategoryRepository
	inserted    *domain.Category
	insertErr   error
	updated     *domain.Category
	reorderSet  []domain.Category
	tourCount   int64
	lockErr     error
	softDeletes int
	sortUpdates map[uuid.UUID]int
}

func (m *mockCategoryRepo) Insert(_ context.Context, _ repository.DB, c *domain.Category) (*domain.Category, error) {
	m.inserted = c
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	c.ID = uuid.MustParse("00000000-0000-0000-0000-00000000000d")
	return c, nil
}
func (m *mockCategoryRepo) Update(_ context.Context, _ repository.DB, c *domain.Category) (*domain.Category, error) {
	m.updated = c
	return c, nil
}
func (m *mockCategoryRepo) LockForUpdate(context.Context, repository.DB, uuid.UUID) error { return m.lockErr }
func (m *mockCategoryRepo) CountActiveToursByCategory(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return m.tourCount, nil
}
func (m *mockCategoryRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error { m.softDeletes++; return nil }
func (m *mockCategoryRepo) ListForReorder(context.Context, repository.DB) ([]domain.Category, error) {
	return m.reorderSet, nil
}
func (m *mockCategoryRepo) UpdateSortOrder(_ context.Context, _ repository.DB, id uuid.UUID, order int) error {
	if m.sortUpdates == nil {
		m.sortUpdates = map[uuid.UUID]int{}
	}
	m.sortUpdates[id] = order
	return nil
}

func newSvc(repo *mockCategoryRepo) (*CategoryService, *fakeTx) {
	tx := &fakeTx{}
	return NewCategoryService(&fakeDB{tx: tx}, repo), tx
}

func TestCreateDerivesSlugAndAppendsAtEnd(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, tx := newSvc(repo)

	c, err := svc.Create(context.Background(), CategoryInput{Name: "Du lịch Đà Nẵng"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.Slug != "du-lich-da-nang" {
		t.Errorf("slug = %q", c.Slug)
	}
	if c.SortOrder != 3 || !c.IsActive {
		t.Errorf("defaults: sort_order=%d (want 3 = appended after 3 rows) is_active=%v", c.SortOrder, c.IsActive)
	}
	if len(repo.sortUpdates) != 0 || !tx.committed {
		t.Errorf("append must not rewrite other ranks (updates=%d) and must commit (%v)", len(repo.sortUpdates), tx.committed)
	}
}

func TestCreateWithPositionRunsRankRewrite(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, _ := newSvc(repo)
	pos := 0
	c, err := svc.Create(context.Background(), CategoryInput{Name: "First", SortOrder: &pos})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.SortOrder != 0 {
		t.Errorf("sort_order = %d, want 0", c.SortOrder)
	}
	// a(1→1 skip), b(2→2 skip), c(3→3 skip) keep their ranks; new row gets 0 and the
	// three seeded rows shift to 1,2,3 — seeded values already equal those ranks.
	if repo.sortUpdates[c.ID] != 0 {
		t.Errorf("expected the new row's rank to be persisted as 0, got %v", repo.sortUpdates)
	}
}

func TestUpdatePreservesRankUnderLockAndMovesOnRequest(t *testing.T) {
	repo := &mockCategoryRepo{reorderSet: seeded()}
	svc, tx := newSvc(repo)
	a := seeded()[0].ID

	u, err := svc.Update(context.Background(), a, CategoryInput{Name: "Renamed", Slug: "renamed"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updated.SortOrder != 1 || u.SortOrder != 1 || !tx.committed {
		t.Errorf("update without position must keep the locked sort_order (got %d) and commit", repo.updated.SortOrder)
	}

	repo = &mockCategoryRepo{reorderSet: seeded()}
	svc, _ = newSvc(repo)
	pos := 2
	u, err = svc.Update(context.Background(), a, CategoryInput{Name: "Renamed", Slug: "renamed", SortOrder: &pos})
	if err != nil {
		t.Fatalf("Update with position: %v", err)
	}
	if u.SortOrder != 2 {
		t.Errorf("sort_order = %d, want 2", u.SortOrder)
	}

	if _, err := svc.Update(context.Background(), uuid.New(), CategoryInput{Name: "X", Slug: "x"}); err == nil {
		t.Error("unknown id must be 404")
	}
}

func TestCreateMapsDuplicateSlugToFieldScoped409(t *testing.T) {
	repo := &mockCategoryRepo{insertErr: repository.CategoryConflict("slug"), reorderSet: seeded()}
	svc, _ := newSvc(repo)

	_, err := svc.Create(context.Background(), CategoryInput{Name: "Island", Slug: "island"})
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeConflict || appErr.Fields["slug"] == "" {
		t.Fatalf("want 409 with fields.slug, got %T %v", err, err)
	}
}

func TestCreateRejectsNonHTTPImageURLServerSide(t *testing.T) {
	svc, _ := newSvc(&mockCategoryRepo{})
	bad := "javascript:alert(1)"
	_, err := svc.Create(context.Background(), CategoryInput{Name: "X", ImageURL: &bad})
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeUnprocessable || appErr.Fields["image_url"] == "" {
		t.Fatalf("want 422 with fields.image_url, got %v", err)
	}
}
