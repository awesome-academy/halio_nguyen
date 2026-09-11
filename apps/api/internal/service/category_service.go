package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// CategoryInput is the explicit create/update DTO (never bind into
// domain.Category — id/created_at/updated_at/deleted_at are not bindable).
// Pointer fields distinguish "omitted" from a zero value. SortOrder is a
// target POSITION (0-based ordinal), never a raw column value — every write
// to sort_order goes through the same locked rank-rewrite (ALG-001) so the
// non-deleted set stays contiguous 0..n-1 (SC-003).
type CategoryInput struct {
	Name        string
	Slug        string
	Description *string
	ImageURL    *string
	SortOrder   *int
	IsActive    *bool
}

// CategoryService owns F002's rules: FR-401 slug derivation, FR-001/402
// uniqueness mapping, BR-001 delete guard (D4) and ALG-001 reorder.
type CategoryService struct {
	db   repository.DB
	repo repository.CategoryRepository
}

// NewCategoryService wires the service against the pool (for transactions)
// and the repository.
func NewCategoryService(db repository.DB, repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{db: db, repo: repo}
}

// List is A1.
func (s *CategoryService) List(ctx context.Context, p repository.CategoryListParams) (*repository.Paginated[repository.CategoryListItem], error) {
	p.Normalize()
	items, total, err := s.repo.List(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &repository.Paginated[repository.CategoryListItem]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// Create is A2. The row is appended at rank n under the set lock, then moved
// to the requested position (if any) by the shared rank-rewrite.
func (s *CategoryService) Create(ctx context.Context, in CategoryInput) (*domain.Category, error) {
	c, err := buildCategory(in)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	cats, err := s.repo.ListForReorder(ctx, tx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	c.SortOrder = len(cats)
	created, err := s.repo.Insert(ctx, tx, c)
	if err != nil {
		return nil, mapCategoryRepoError(err)
	}
	if in.SortOrder != nil {
		ordered, err := s.rewriteRanks(ctx, tx, append(cats, *created), created.ID, *in.SortOrder)
		if err != nil {
			return nil, err
		}
		created.SortOrder = ordered[indexOfCategory(ordered, created.ID)].SortOrder
	}
	return created, commit(ctx, tx)
}

// Update is A3 — a full replace. The current sort_order is read under the
// same set lock the write runs in (no unlocked read → no lost reorder), and
// a requested position is applied through the rank-rewrite.
func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, in CategoryInput) (*domain.Category, error) {
	c, err := buildCategory(in)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	cats, err := s.repo.ListForReorder(ctx, tx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	cur := indexOfCategory(cats, id)
	if cur < 0 {
		return nil, apperror.NewNotFound("Category not found")
	}
	c.ID = id
	c.SortOrder = cats[cur].SortOrder
	updated, err := s.repo.Update(ctx, tx, c)
	if err != nil {
		return nil, mapCategoryRepoError(err)
	}
	if in.SortOrder != nil {
		cats[cur] = *updated
		ordered, err := s.rewriteRanks(ctx, tx, cats, id, *in.SortOrder)
		if err != nil {
			return nil, err
		}
		updated.SortOrder = ordered[indexOfCategory(ordered, id)].SortOrder
	}
	return updated, commit(ctx, tx)
}

// Delete is A5 (BR-001 / D4): lock the row, count live tours, and either
// refuse with 409 (zero writes) or soft-delete — count and write share one
// transaction so a concurrent tour-create cannot slip between them.
func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.repo.LockForUpdate(ctx, tx, id); err != nil {
		return mapCategoryRepoError(err)
	}
	n, err := s.repo.CountActiveToursByCategory(ctx, tx, id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if n > 0 {
		return apperror.NewConflict(
			fmt.Sprintf("Cannot delete: %d tour(s) still use this category", n),
			map[string]string{"tour_count": fmt.Sprintf("%d", n)},
		)
	}
	if err := s.repo.SoftDelete(ctx, tx, id); err != nil {
		return mapCategoryRepoError(err)
	}
	return commit(ctx, tx)
}

func commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		return apperror.NewInternal(err)
	}
	return nil
}

func mapCategoryRepoError(err error) error {
	if errors.Is(err, repository.ErrCategoryNotFound) {
		return apperror.NewNotFound("Category not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
