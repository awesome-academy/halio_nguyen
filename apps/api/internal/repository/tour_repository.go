package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrTourNotFound is returned when no non-deleted tour matches.
var ErrTourNotFound = errors.New("repository: tour not found")

// TourListParams adds F003's catalog filters to the shared list params.
type TourListParams struct {
	ListParams
	CategoryID *uuid.UUID
	Status     *string
	PriceMin   *float64
	PriceMax   *float64
}

// TourRepository is the pgx data access A1-A6 depend on. Methods take the
// shared DB so the service can hand them an open pgx.Tx (A3 create, A6
// delete guard).
type TourRepository interface {
	List(ctx context.Context, db DB, p TourListParams) ([]domain.Tour, int64, error)
	FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Tour, error)
	Insert(ctx context.Context, db DB, t *domain.Tour) (*domain.Tour, error)
	Update(ctx context.Context, db DB, t *domain.Tour) (*domain.Tour, error)
	UpdateStatus(ctx context.Context, db DB, id uuid.UUID, status string) (*domain.Tour, error)
	SoftDelete(ctx context.Context, db DB, id uuid.UUID) error
	// LockForUpdate takes a row lock on a non-deleted tour so a concurrent
	// booking-create cannot slip between BR-012's count and the soft-delete
	// (A6 TOCTOU) — mirrors CategoryRepository.LockForUpdate (phase-03).
	// Implemented in tour_repository_tx.go, alongside UpdateStatus/SoftDelete,
	// to keep this file under the 200-line cap.
	LockForUpdate(ctx context.Context, db DB, id uuid.UUID) error
}

// tourSortAllow maps sort_by values to ORDER BY literals (BR-008); "" is the
// default. SC-001 requires an unrecognised sort_by to fall back to
// created_at DESC — not just the column but the direction too — so List
// defaults SortDir to desc before resolving when the caller left it unset.
var tourSortAllow = map[string]string{
	"":            "t.created_at",
	"title":       "t.title",
	"price":       "t.price",
	"created_at":  "t.created_at",
	"destination": "t.destination",
}

// tourUniqueConstraints maps tours' UNIQUE constraint to the request field it
// belongs to (inline UNIQUE on tours.slug auto-names to tours_slug_key).
var tourUniqueConstraints = map[string]string{
	"tours_slug_key": "slug",
}

const tourColumns = `t.id, t.category_id, t.title, t.slug, t.description, t.itinerary, t.destination,
	t.duration_days, t.duration_nights, t.price, t.discount_price, t.max_participants, t.thumbnail_url,
	t.highlights, t.inclusions, t.exclusions, t.status, t.avg_rating, t.total_ratings,
	t.created_at, t.updated_at, t.deleted_at`

type tourRepository struct{}

// NewTourRepository returns the pgx-backed TourRepository.
func NewTourRepository() TourRepository {
	return tourRepository{}
}

func (tourRepository) List(ctx context.Context, db DB, p TourListParams) ([]domain.Tour, int64, error) {
	p.Normalize()
	sortDir := p.SortDir
	if sortDir == "" {
		sortDir = SortDesc
	}
	col, dir := ResolveSort(tourSortAllow, p.SortBy, sortDir)

	var search *string
	if p.Search != "" {
		search = &p.Search
	}

	rows, err := db.Query(ctx,
		`SELECT `+tourColumns+`, COUNT(*) OVER() AS total_count
		 FROM tours t
		 WHERE t.deleted_at IS NULL
		   AND (@search::text IS NULL OR t.title ILIKE '%' || @search || '%' OR t.destination ILIKE '%' || @search || '%')
		   AND (@category_id::uuid IS NULL OR t.category_id = @category_id)
		   AND (@status::text IS NULL OR t.status = @status)
		   AND (@price_min::numeric IS NULL OR t.price >= @price_min)
		   AND (@price_max::numeric IS NULL OR t.price <= @price_max)
		 ORDER BY `+col+` `+dir+`, t.created_at, t.id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{
			"search": search, "category_id": p.CategoryID, "status": p.Status,
			"price_min": p.PriceMin, "price_max": p.PriceMax,
			"limit": p.PageSize, "offset": p.Offset(),
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing tours: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Tour, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var t domain.Tour
		var highlights []string
		if err := rows.Scan(
			&t.ID, &t.CategoryID, &t.Title, &t.Slug, &t.Description, &t.Itinerary, &t.Destination,
			&t.DurationDays, &t.DurationNights, &t.Price, &t.DiscountPrice, &t.MaxParticipants, &t.ThumbnailURL,
			&highlights, &t.Inclusions, &t.Exclusions, &t.Status, &t.AvgRating, &t.TotalRatings,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning tour: %w", err)
		}
		t.Highlights = highlights
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (tourRepository) FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Tour, error) {
	row := db.QueryRow(ctx,
		`SELECT `+tourColumns+` FROM tours t WHERE t.id = @id AND t.deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	return scanTour(row)
}

func (tourRepository) Insert(ctx context.Context, db DB, t *domain.Tour) (*domain.Tour, error) {
	row := db.QueryRow(ctx,
		`INSERT INTO tours AS t (category_id, title, slug, description, itinerary, destination,
			duration_days, duration_nights, price, discount_price, max_participants, thumbnail_url,
			highlights, inclusions, exclusions, status)
		 VALUES (@category_id, @title, @slug, @description, @itinerary, @destination,
			@duration_days, @duration_nights, @price, @discount_price, @max_participants, @thumbnail_url,
			@highlights, @inclusions, @exclusions, @status)
		 RETURNING `+tourColumns,
		tourArgs(t),
	)
	return scanTour(row)
}

// Update writes every A4-editable column except status: FR-305 makes A5 the
// single writer of tours.status (see tour_service.go's judgment-call note),
// so status is deliberately left out of this SET clause.
func (tourRepository) Update(ctx context.Context, db DB, t *domain.Tour) (*domain.Tour, error) {
	args := tourArgs(t)
	args["id"] = t.ID
	row := db.QueryRow(ctx,
		`UPDATE tours AS t
		 SET category_id = @category_id, title = @title, slug = @slug, description = @description,
		     itinerary = @itinerary, destination = @destination, duration_days = @duration_days,
		     duration_nights = @duration_nights, price = @price, discount_price = @discount_price,
		     max_participants = @max_participants, thumbnail_url = @thumbnail_url, highlights = @highlights,
		     inclusions = @inclusions, exclusions = @exclusions
		 WHERE t.id = @id AND t.deleted_at IS NULL
		 RETURNING `+tourColumns,
		args,
	)
	return scanTour(row)
}

func tourArgs(t *domain.Tour) pgx.NamedArgs {
	return pgx.NamedArgs{
		"category_id": t.CategoryID, "title": t.Title, "slug": t.Slug, "description": t.Description,
		"itinerary": t.Itinerary, "destination": t.Destination, "duration_days": t.DurationDays,
		"duration_nights": t.DurationNights, "price": t.Price, "discount_price": t.DiscountPrice,
		"max_participants": t.MaxParticipants, "thumbnail_url": t.ThumbnailURL, "highlights": t.Highlights,
		"inclusions": t.Inclusions, "exclusions": t.Exclusions, "status": t.Status,
	}
}
