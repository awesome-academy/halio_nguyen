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

// The lock/soft-delete pair in this file runs inside the open pgx.Tx A13's
// BR-012 guard opens (mirrors tour_repository_tx.go).

func (tourScheduleRepository) UpdateStatus(ctx context.Context, db DB, tourID, id uuid.UUID, status string) (*domain.TourSchedule, error) {
	row := db.QueryRow(ctx,
		`UPDATE tour_schedules AS ts SET status = @status
		 WHERE ts.id = @id AND ts.tour_id = @tour_id AND ts.deleted_at IS NULL
		 RETURNING `+tourScheduleColumns,
		pgx.NamedArgs{"id": id, "tour_id": tourID, "status": status},
	)
	return scanTourScheduleRow(row)
}

func (tourScheduleRepository) SoftDelete(ctx context.Context, db DB, tourID, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE tour_schedules SET deleted_at = NOW() WHERE id = @id AND tour_id = @tour_id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": id, "tour_id": tourID},
	)
	if err != nil {
		return fmt.Errorf("repository: soft-deleting tour schedule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTourScheduleNotFound
	}
	return nil
}

func (tourScheduleRepository) LockForUpdate(ctx context.Context, db DB, tourID, id uuid.UUID) error {
	var got uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM tour_schedules WHERE id = @id AND tour_id = @tour_id AND deleted_at IS NULL FOR UPDATE`,
		pgx.NamedArgs{"id": id, "tour_id": tourID},
	).Scan(&got)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTourScheduleNotFound
	}
	if err != nil {
		return fmt.Errorf("repository: locking tour schedule: %w", err)
	}
	return nil
}

// RestoreSlots is phase-06's BR-005 write: a cancelled booking hands its
// seats back to the schedule. The addition happens in SQL rather than
// read-modify-write in Go so two concurrent cancels on the same schedule
// cannot lose one another's increment. A missing schedule is a hard error —
// silently dropping it would leak the seats permanently, and no other admin
// path writes available_slots.
func (tourScheduleRepository) RestoreSlots(ctx context.Context, db DB, scheduleID uuid.UUID, n int) error {
	tag, err := db.Exec(ctx,
		`UPDATE tour_schedules SET available_slots = available_slots + @n, updated_at = NOW()
		 WHERE id = @id AND deleted_at IS NULL`,
		pgx.NamedArgs{"id": scheduleID, "n": n},
	)
	if err != nil {
		return fmt.Errorf("repository: restoring schedule slots: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTourScheduleNotFound
	}
	return nil
}

// ScheduleDateConflict builds the 409 for a duplicate departure_date (BR-004).
func ScheduleDateConflict(field string) *apperror.Error {
	msg := "A schedule already exists for this departure date."
	return apperror.NewConflict(msg, map[string]string{field: msg})
}
