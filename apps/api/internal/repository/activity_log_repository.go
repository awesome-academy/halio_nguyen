package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ActivityLogRepository inserts audit rows. It takes the shared DB
// interface (not *pgxpool.Pool) so a service can enlist the insert in an
// open transaction — Phase 6 reuses this same method for cancel_tour.
type ActivityLogRepository interface {
	Insert(ctx context.Context, db DB, log NewActivityLog) error
}

// NewActivityLog is the write-side shape for one activity_logs row.
// activity_logs.user_id and entity_id are both NOT NULL (see decisions.md
// L1) — there is no path to log an event with no identified subject.
type NewActivityLog struct {
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Metadata   json.RawMessage
	IPAddress  string
	UserAgent  string
}

type activityLogRepository struct{}

// NewActivityLogRepository returns the pgx-backed ActivityLogRepository.
func NewActivityLogRepository() ActivityLogRepository {
	return activityLogRepository{}
}

func (activityLogRepository) Insert(ctx context.Context, db DB, log NewActivityLog) error {
	// ip_address is INET; an empty string is not a valid CIDR literal, so it
	// must go in as SQL NULL rather than "".
	var ipAddress any
	if log.IPAddress != "" {
		ipAddress = log.IPAddress
	}

	_, err := db.Exec(ctx,
		`INSERT INTO activity_logs (user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
		 VALUES (@user_id, @action, @entity_type, @entity_id, @metadata, @ip_address, @user_agent)`,
		pgx.NamedArgs{
			"user_id":     log.UserID,
			"action":      log.Action,
			"entity_type": log.EntityType,
			"entity_id":   log.EntityID,
			"metadata":    log.Metadata,
			"ip_address":  ipAddress,
			"user_agent":  log.UserAgent,
		},
	)
	if err != nil {
		return fmt.Errorf("repository: inserting activity log: %w", err)
	}
	return nil
}
