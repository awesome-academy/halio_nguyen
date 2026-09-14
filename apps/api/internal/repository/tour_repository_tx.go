package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// The methods in this file are the ones TourService calls inside an open
// pgx.Tx (A6 delete guard); db is expected to be that tx there. UpdateStatus
// runs against the plain pool (A5 has no transaction — see tour_service.go).

func (tourRepository) UpdateStatus(ctx context.Context, db DB, id uuid.UUID, status string) (*domain.Tour, error) {
	row := db.QueryRow(ctx,
		`UPDATE tours AS t SET status = @status WHERE t.id = @id AND t.deleted_at IS NULL RETURNING `+tourColumns,
		pgx.NamedArgs{"id": id, "status": status},
	)
	return scanTour(row)
}

func (tourRepository) SoftDelete(ctx context.Context, db DB, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE tours SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return fmt.Errorf("repository: soft-deleting tour: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTourNotFound
	}
	return nil
}

func (tourRepository) LockForUpdate(ctx context.Context, db DB, id uuid.UUID) error {
	var got uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM tours WHERE id = @id AND deleted_at IS NULL FOR UPDATE`,
		pgx.NamedArgs{"id": id},
	).Scan(&got)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTourNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: locking tour: %w", err)
	}
	return nil
}

// scanTour maps one row and translates the two failure modes every caller
// shares: no row -> ErrTourNotFound, unique violation -> a 409 naming the
// colliding field (edge case table: "A tour with this title already exists.").
func scanTour(row pgx.Row) (*domain.Tour, error) {
	var t domain.Tour
	var highlights []string
	err := row.Scan(
		&t.ID, &t.CategoryID, &t.Title, &t.Slug, &t.Description, &t.Itinerary, &t.Destination,
		&t.DurationDays, &t.DurationNights, &t.Price, &t.DiscountPrice, &t.MaxParticipants, &t.ThumbnailURL,
		&highlights, &t.Inclusions, &t.Exclusions, &t.Status, &t.AvgRating, &t.TotalRatings,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTourNotFound
	}
	if field, ok := UniqueViolationField(err, tourUniqueConstraints); ok {
		return nil, TourConflict(field)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning tour: %w", err)
	}
	t.Highlights = highlights
	return &t, nil
}

// TourConflict builds the 409 for a duplicate slug (edge case table).
func TourConflict(field string) *apperror.Error {
	msg := "A tour with this title already exists."
	return apperror.NewConflict(msg, map[string]string{field: msg})
}
