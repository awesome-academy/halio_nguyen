package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// The methods in this file are the ones CategoryService calls inside an
// open pgx.Tx (A4 reorder, A5 delete); db is expected to be that tx.

func (categoryRepository) LockForUpdate(ctx context.Context, db DB, id uuid.UUID) error {
	var got uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM categories WHERE id = @id AND deleted_at IS NULL FOR UPDATE`,
		pgx.NamedArgs{"id": id},
	).Scan(&got)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCategoryNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: locking category: %w", err)
	}
	return nil
}

func (categoryRepository) CountActiveToursByCategory(ctx context.Context, db DB, id uuid.UUID) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM tours WHERE category_id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("repository: counting tours for category: %w", err)
	}
	return n, nil
}

func (categoryRepository) SoftDelete(ctx context.Context, db DB, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE categories SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return fmt.Errorf("repository: soft-deleting category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (categoryRepository) ListForReorder(ctx context.Context, db DB) ([]domain.Category, error) {
	rows, err := db.Query(ctx,
		`SELECT `+categoryColumns+` FROM categories c
		 WHERE c.deleted_at IS NULL ORDER BY c.sort_order, c.created_at, c.id FOR UPDATE`,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing categories for reorder: %w", err)
	}
	defer rows.Close()

	out := []domain.Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (categoryRepository) UpdateSortOrder(ctx context.Context, db DB, id uuid.UUID, sortOrder int) error {
	_, err := db.Exec(ctx,
		`UPDATE categories SET sort_order = @sort_order WHERE id = @id`,
		pgx.NamedArgs{"id": id, "sort_order": sortOrder},
	)
	if err != nil {
		return fmt.Errorf("repository: updating sort_order: %w", err)
	}
	return nil
}
