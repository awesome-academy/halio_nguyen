package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// Reorder is A4 (ALG-001, revised): positions are ordinal ranks derived
// from the locked, ordered non-deleted set — never the stored sort_order —
// so pre-existing gaps/duplicates (seed starts at 1) cannot cause an
// off-by-one. The whole set is rewritten to 0..n-1 in one transaction and
// returned so the UI needs no refetch.
func (s *CategoryService) Reorder(ctx context.Context, id uuid.UUID, target int) ([]domain.Category, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	cats, err := s.repo.ListForReorder(ctx, tx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	ordered, err := s.rewriteRanks(ctx, tx, cats, id, target)
	if err != nil {
		return nil, err
	}
	return ordered, commit(ctx, tx)
}

// rewriteRanks moves id to the clamped target within the already-locked set
// and persists every rank that changed. It is the single write path for
// sort_order (Create/Update/Reorder all funnel through it), which is what
// keeps SC-003 (contiguous 0..n-1, no duplicates) true after any write.
func (s *CategoryService) rewriteRanks(ctx context.Context, tx pgx.Tx, cats []domain.Category, id uuid.UUID, target int) ([]domain.Category, error) {
	cur := indexOfCategory(cats, id)
	if cur < 0 {
		return nil, apperror.NewNotFound("Category not found")
	}
	ordered := moveCategory(cats, cur, clampRank(target, len(cats)))
	for rank := range ordered {
		if ordered[rank].SortOrder == rank {
			continue
		}
		ordered[rank].SortOrder = rank
		if err := s.repo.UpdateSortOrder(ctx, tx, ordered[rank].ID, rank); err != nil {
			return nil, apperror.NewInternal(err)
		}
	}
	return ordered, nil
}
