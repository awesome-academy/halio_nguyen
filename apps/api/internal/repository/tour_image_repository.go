package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrTourImageNotFound is returned when no non-deleted image matches the
// given (tour_id, id) pair.
var ErrTourImageNotFound = errors.New("repository: tour image not found")

// TourImageRepository is the pgx data access A7-A9 depend on. Every method
// scopes by both tour_id and id so an image cannot be moved between tours by
// swapping ids (IDOR — phase-04 Security Considerations).
type TourImageRepository interface {
	// ListByTour returns every non-deleted image for a non-deleted tour,
	// ordered by sort_order (Key Insight: child reads must also filter the
	// parent's deleted_at).
	ListByTour(ctx context.Context, db DB, tourID uuid.UUID) ([]domain.TourImage, error)
	Insert(ctx context.Context, db DB, img *domain.TourImage) (*domain.TourImage, error)
	Update(ctx context.Context, db DB, tourID, id uuid.UUID, caption *string, sortOrder int) (*domain.TourImage, error)
	SoftDelete(ctx context.Context, db DB, tourID, id uuid.UUID) error
	// NextSortOrder is current max(sort_order)+1 for the tour, or 0 when it
	// has no images yet (A7's "next available position").
	NextSortOrder(ctx context.Context, db DB, tourID uuid.UUID) (int, error)
}

const tourImageColumns = `ti.id, ti.tour_id, ti.image_url, ti.caption, ti.sort_order, ti.created_at, ti.deleted_at`

type tourImageRepository struct{}

// NewTourImageRepository returns the pgx-backed TourImageRepository.
func NewTourImageRepository() TourImageRepository {
	return tourImageRepository{}
}

func (tourImageRepository) ListByTour(ctx context.Context, db DB, tourID uuid.UUID) ([]domain.TourImage, error) {
	rows, err := db.Query(ctx,
		`SELECT `+tourImageColumns+`
		 FROM tour_images ti
		 JOIN tours t ON t.id = ti.tour_id AND t.deleted_at IS NULL
		 WHERE ti.tour_id = @tour_id AND ti.deleted_at IS NULL
		 ORDER BY ti.sort_order, ti.created_at, ti.id`,
		pgx.NamedArgs{"tour_id": tourID},
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing tour images: %w", err)
	}
	defer rows.Close()

	items := []domain.TourImage{}
	for rows.Next() {
		img, err := scanTourImage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *img)
	}
	return items, rows.Err()
}

func (tourImageRepository) Insert(ctx context.Context, db DB, img *domain.TourImage) (*domain.TourImage, error) {
	row := db.QueryRow(ctx,
		`INSERT INTO tour_images AS ti (tour_id, image_url, caption, sort_order)
		 VALUES (@tour_id, @image_url, @caption, @sort_order)
		 RETURNING `+tourImageColumns,
		pgx.NamedArgs{
			"tour_id": img.TourID, "image_url": img.ImageURL, "caption": img.Caption, "sort_order": img.SortOrder,
		},
	)
	return scanTourImageRow(row)
}

// Update writes caption and sort_order verbatim (BR-011 — no renumbering, no
// gap-filling of sibling rows).
func (tourImageRepository) Update(ctx context.Context, db DB, tourID, id uuid.UUID, caption *string, sortOrder int) (*domain.TourImage, error) {
	row := db.QueryRow(ctx,
		`UPDATE tour_images AS ti SET caption = @caption, sort_order = @sort_order
		 WHERE ti.id = @id AND ti.tour_id = @tour_id AND ti.deleted_at IS NULL
		 RETURNING `+tourImageColumns,
		pgx.NamedArgs{"id": id, "tour_id": tourID, "caption": caption, "sort_order": sortOrder},
	)
	return scanTourImageRow(row)
}

func (tourImageRepository) SoftDelete(ctx context.Context, db DB, tourID, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE tour_images SET deleted_at = NOW() WHERE id = @id AND tour_id = @tour_id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id, "tour_id": tourID},
	)
	if err != nil {
		return fmt.Errorf("repository: soft-deleting tour image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTourImageNotFound
	}
	return nil
}

func (tourImageRepository) NextSortOrder(ctx context.Context, db DB, tourID uuid.UUID) (int, error) {
	var next int
	err := db.QueryRow(ctx,
		`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM tour_images WHERE tour_id = @tour_id AND deleted_at IS NULL`,
		pgx.NamedArgs{"tour_id": tourID},
	).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("repository: computing next sort_order: %w", err)
	}
	return next, nil
}

// scanTourImage scans from pgx.Rows (List); scanTourImageRow scans from a
// single pgx.Row (Insert/Update). Both share the same failure mapping: no
// row -> ErrTourImageNotFound.
func scanTourImage(rows pgx.Rows) (*domain.TourImage, error) {
	var img domain.TourImage
	if err := rows.Scan(&img.ID, &img.TourID, &img.ImageURL, &img.Caption, &img.SortOrder, &img.CreatedAt, &img.DeletedAt); err != nil {
		return nil, fmt.Errorf("repository: scanning tour image: %w", err)
	}
	return &img, nil
}

func scanTourImageRow(row pgx.Row) (*domain.TourImage, error) {
	var img domain.TourImage
	err := row.Scan(&img.ID, &img.TourID, &img.ImageURL, &img.Caption, &img.SortOrder, &img.CreatedAt, &img.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTourImageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning tour image: %w", err)
	}
	return &img, nil
}
