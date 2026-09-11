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

// ErrCategoryNotFound is returned when no non-deleted category matches.
var ErrCategoryNotFound = errors.New("repository: category not found")

// CategoryListParams adds F002's is_active filter to the shared list params.
type CategoryListParams struct {
	ListParams
	IsActive *bool
}

// CategoryListItem is one A1 row: the category plus its derived tour_count.
type CategoryListItem struct {
	domain.Category
	TourCount int64 `json:"tour_count"`
}

// CategoryRepository is the pgx data access F002 depends on. Methods take
// the shared DB so the service can hand them an open pgx.Tx (reorder and
// delete run inside one transaction).
type CategoryRepository interface {
	List(ctx context.Context, db DB, p CategoryListParams) ([]CategoryListItem, int64, error)
	FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Category, error)
	Insert(ctx context.Context, db DB, c *domain.Category) (*domain.Category, error)
	Update(ctx context.Context, db DB, c *domain.Category) (*domain.Category, error)
	// LockForUpdate takes a row lock on a non-deleted category so a
	// concurrent tour-create cannot slip between the BR-001 count and the
	// soft-delete (A5 TOCTOU). Returns ErrCategoryNotFound when absent.
	LockForUpdate(ctx context.Context, db DB, id uuid.UUID) error
	CountActiveToursByCategory(ctx context.Context, db DB, id uuid.UUID) (int64, error)
	SoftDelete(ctx context.Context, db DB, id uuid.UUID) error
	// ListForReorder returns every non-deleted category in display order,
	// row-locked, so ALG-001 can rewrite ranks without interleaving.
	ListForReorder(ctx context.Context, db DB) ([]domain.Category, error)
	UpdateSortOrder(ctx context.Context, db DB, id uuid.UUID, sortOrder int) error
}

// categorySortAllow maps sort_by values to ORDER BY literals; "" is the
// default (FR-201: sort_order ASC).
var categorySortAllow = map[string]string{
	"":           "c.sort_order",
	"name":       "c.name",
	"sort_order": "c.sort_order",
	"created_at": "c.created_at",
}

// categoryUniqueConstraints maps the schema's UNIQUE constraints to the
// request field they belong to (categories.name / categories.slug are
// table-level UNIQUE — see phase-03 Key Insights on why deleted rows still
// collide).
var categoryUniqueConstraints = map[string]string{
	"categories_name_key": "name",
	"categories_slug_key": "slug",
}

const categoryColumns = `c.id, c.name, c.slug, c.description, c.image_url, c.sort_order, c.is_active, c.created_at, c.updated_at, c.deleted_at`

type categoryRepository struct{}

// NewCategoryRepository returns the pgx-backed CategoryRepository.
func NewCategoryRepository() CategoryRepository {
	return categoryRepository{}
}

func (categoryRepository) List(ctx context.Context, db DB, p CategoryListParams) ([]CategoryListItem, int64, error) {
	p.Normalize()
	col, dir := ResolveSort(categorySortAllow, p.SortBy, p.SortDir)

	var search *string
	if p.Search != "" {
		search = &p.Search
	}

	rows, err := db.Query(ctx,
		`SELECT `+categoryColumns+`, COALESCE(t.cnt, 0) AS tour_count, COUNT(*) OVER() AS total_count
		 FROM categories c
		 LEFT JOIN (SELECT category_id, COUNT(*) AS cnt FROM tours WHERE deleted_at IS NULL GROUP BY category_id) t
		   ON t.category_id = c.id
		 WHERE c.deleted_at IS NULL
		   AND (@search::text IS NULL OR c.name ILIKE '%' || @search || '%')
		   AND (@is_active::bool IS NULL OR c.is_active = @is_active)
		 ORDER BY `+col+` `+dir+`, c.created_at, c.id
		 LIMIT @limit OFFSET @offset`,
		pgx.NamedArgs{"search": search, "is_active": p.IsActive, "limit": p.PageSize, "offset": p.Offset()},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: listing categories: %w", err)
	}
	defer rows.Close()

	items := make([]CategoryListItem, 0, p.PageSize)
	var total int64
	for rows.Next() {
		var it CategoryListItem
		if err := rows.Scan(
			&it.ID, &it.Name, &it.Slug, &it.Description, &it.ImageURL, &it.SortOrder, &it.IsActive,
			&it.CreatedAt, &it.UpdatedAt, &it.DeletedAt, &it.TourCount, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("repository: scanning category: %w", err)
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

func (categoryRepository) FindByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Category, error) {
	row := db.QueryRow(ctx,
		`SELECT `+categoryColumns+` FROM categories c WHERE c.id = @id AND c.deleted_at IS NULL`,
		pgx.NamedArgs{"id": id},
	)
	return scanCategory(row)
}

func (categoryRepository) Insert(ctx context.Context, db DB, c *domain.Category) (*domain.Category, error) {
	row := db.QueryRow(ctx,
		`INSERT INTO categories AS c (name, slug, description, image_url, sort_order, is_active)
		 VALUES (@name, @slug, @description, @image_url, @sort_order, @is_active)
		 RETURNING `+categoryColumns,
		categoryArgs(c),
	)
	return scanCategory(row)
}

func (categoryRepository) Update(ctx context.Context, db DB, c *domain.Category) (*domain.Category, error) {
	args := categoryArgs(c)
	args["id"] = c.ID
	row := db.QueryRow(ctx,
		`UPDATE categories AS c
		 SET name = @name, slug = @slug, description = @description, image_url = @image_url,
		     sort_order = @sort_order, is_active = @is_active
		 WHERE c.id = @id AND c.deleted_at IS NULL
		 RETURNING `+categoryColumns,
		args,
	)
	return scanCategory(row)
}

func categoryArgs(c *domain.Category) pgx.NamedArgs {
	return pgx.NamedArgs{
		"name": c.Name, "slug": c.Slug, "description": c.Description, "image_url": c.ImageURL,
		"sort_order": c.SortOrder, "is_active": c.IsActive,
	}
}

// scanCategory maps one row and translates the two failure modes every
// caller shares: no row → ErrCategoryNotFound, unique violation → a 409
// naming the colliding field (FR-402 / SC-001). Never the constraint name.
func scanCategory(row pgx.Row) (*domain.Category, error) {
	var c domain.Category
	err := row.Scan(
		&c.ID, &c.Name, &c.Slug, &c.Description, &c.ImageURL, &c.SortOrder, &c.IsActive,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if field, ok := UniqueViolationField(err, categoryUniqueConstraints); ok {
		return nil, CategoryConflict(field)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning category: %w", err)
	}
	return &c, nil
}

// CategoryConflict builds the FR-402 409 for a duplicate name/slug. The
// message says "including deleted ones" because the UNIQUE constraints are
// table-level, so a soft-deleted category keeps its name and slug reserved
// (limitation L8).
func CategoryConflict(field string) *apperror.Error {
	msg := fmt.Sprintf("A category with this %s already exists, including deleted ones", field)
	return apperror.NewConflict(msg, map[string]string{field: "This " + field + " is already in use"})
}
