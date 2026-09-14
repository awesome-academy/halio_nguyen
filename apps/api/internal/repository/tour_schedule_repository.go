package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
)

// ErrTourScheduleNotFound is returned when no non-deleted schedule matches
// the given (tour_id, id) pair.
var ErrTourScheduleNotFound = errors.New("repository: tour schedule not found")

// tourScheduleUniqueConstraints maps the schema's UNIQUE constraint to the
// request field it belongs to (BR-004).
var tourScheduleUniqueConstraints = map[string]string{
	"uq_tour_schedule_date": "departure_date",
}

const tourScheduleColumns = `ts.id, ts.tour_id, ts.departure_date, ts.return_date, ts.available_slots,
	ts.price_override, ts.status, ts.created_at, ts.updated_at, ts.deleted_at`

// TourScheduleRepository is the pgx data access A10-A13 depend on. Every
// method taking an id scopes by both tour_id and id (IDOR — phase-04
// Security Considerations).
type TourScheduleRepository interface {
	// ListByTour returns every non-deleted schedule for a non-deleted tour
	// (Key Insight: child reads must also filter the parent's deleted_at).
	ListByTour(ctx context.Context, db DB, tourID uuid.UUID) ([]domain.TourSchedule, error)
	FindByID(ctx context.Context, db DB, tourID, id uuid.UUID) (*domain.TourSchedule, error)
	Insert(ctx context.Context, db DB, s *domain.TourSchedule) (*domain.TourSchedule, error)
	Update(ctx context.Context, db DB, s *domain.TourSchedule) (*domain.TourSchedule, error)
	UpdateStatus(ctx context.Context, db DB, tourID, id uuid.UUID, status string) (*domain.TourSchedule, error)
	SoftDelete(ctx context.Context, db DB, tourID, id uuid.UUID) error
	// LockForUpdate mirrors TourRepository.LockForUpdate for BR-012's A13
	// guard (implemented in tour_schedule_repository_tx.go).
	LockForUpdate(ctx context.Context, db DB, tourID, id uuid.UUID) error
}

type tourScheduleRepository struct{}

// NewTourScheduleRepository returns the pgx-backed TourScheduleRepository.
func NewTourScheduleRepository() TourScheduleRepository {
	return tourScheduleRepository{}
}

func (tourScheduleRepository) ListByTour(ctx context.Context, db DB, tourID uuid.UUID) ([]domain.TourSchedule, error) {
	rows, err := db.Query(ctx,
		`SELECT `+tourScheduleColumns+`
		 FROM tour_schedules ts
		 JOIN tours t ON t.id = ts.tour_id AND t.deleted_at IS NULL
		 WHERE ts.tour_id = @tour_id AND ts.deleted_at IS NULL
		 ORDER BY ts.departure_date, ts.id`,
		pgx.NamedArgs{"tour_id": tourID},
	)
	if err != nil {
		return nil, fmt.Errorf("repository: listing tour schedules: %w", err)
	}
	defer rows.Close()

	items := []domain.TourSchedule{}
	for rows.Next() {
		s, err := scanTourSchedule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

func (tourScheduleRepository) FindByID(ctx context.Context, db DB, tourID, id uuid.UUID) (*domain.TourSchedule, error) {
	row := db.QueryRow(ctx,
		`SELECT `+tourScheduleColumns+`
		 FROM tour_schedules ts
		 JOIN tours t ON t.id = ts.tour_id AND t.deleted_at IS NULL
		 WHERE ts.id = @id AND ts.tour_id = @tour_id AND ts.deleted_at IS NULL`,
		pgx.NamedArgs{"id": id, "tour_id": tourID},
	)
	return scanTourScheduleRow(row)
}

func (tourScheduleRepository) Insert(ctx context.Context, db DB, s *domain.TourSchedule) (*domain.TourSchedule, error) {
	row := db.QueryRow(ctx,
		`INSERT INTO tour_schedules AS ts (tour_id, departure_date, return_date, available_slots, price_override, status)
		 VALUES (@tour_id, @departure_date, @return_date, @available_slots, @price_override, @status)
		 RETURNING `+tourScheduleColumns,
		tourScheduleArgs(s),
	)
	return scanTourScheduleRow(row)
}

// Update writes every A11-editable column except status: A12 is the single
// writer of tour_schedules.status (mirrors tours/A5).
func (tourScheduleRepository) Update(ctx context.Context, db DB, s *domain.TourSchedule) (*domain.TourSchedule, error) {
	args := tourScheduleArgs(s)
	args["id"] = s.ID
	row := db.QueryRow(ctx,
		`UPDATE tour_schedules AS ts
		 SET departure_date = @departure_date, return_date = @return_date,
		     available_slots = @available_slots, price_override = @price_override
		 WHERE ts.id = @id AND ts.tour_id = @tour_id AND ts.deleted_at IS NULL
		 RETURNING `+tourScheduleColumns,
		args,
	)
	return scanTourScheduleRow(row)
}

func tourScheduleArgs(s *domain.TourSchedule) pgx.NamedArgs {
	return pgx.NamedArgs{
		"tour_id": s.TourID, "departure_date": s.DepartureDate, "return_date": s.ReturnDate,
		"available_slots": s.AvailableSlots, "price_override": s.PriceOverride, "status": s.Status,
	}
}

// scanTourSchedule scans from pgx.Rows (ListByTour); scanTourScheduleRow
// scans from a single pgx.Row (FindByID/Insert/Update) and additionally maps
// no-row -> ErrTourScheduleNotFound and a unique violation -> BR-004's 409.
func scanTourSchedule(rows pgx.Rows) (*domain.TourSchedule, error) {
	var s domain.TourSchedule
	if err := rows.Scan(&s.ID, &s.TourID, &s.DepartureDate, &s.ReturnDate, &s.AvailableSlots, &s.PriceOverride, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
		return nil, fmt.Errorf("repository: scanning tour schedule: %w", err)
	}
	return &s, nil
}

func scanTourScheduleRow(row pgx.Row) (*domain.TourSchedule, error) {
	var s domain.TourSchedule
	err := row.Scan(&s.ID, &s.TourID, &s.DepartureDate, &s.ReturnDate, &s.AvailableSlots, &s.PriceOverride, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTourScheduleNotFound
	}
	if field, ok := UniqueViolationField(err, tourScheduleUniqueConstraints); ok {
		return nil, ScheduleDateConflict(field)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: scanning tour schedule: %w", err)
	}
	return &s, nil
}
