package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Activity action constants per requirement:
// - Booking tour
// - Cancel tour
// - Create new review
// - User pay tour
const (
	ActionBookingTour   = "booking_tour"
	ActionCancelTour    = "cancel_tour"
	ActionCreateReview  = "create_review"
	ActionPayTour       = "pay_tour"
	ActionRateTour      = "rate_tour"
	ActionCommentReview = "comment_review"
	ActionLikeReview    = "like_review"
	ActionLogin         = "login"
	ActionLogout        = "logout"
)

// Entity types
const (
	EntityTypeBooking = "booking"
	EntityTypeTour    = "tour"
	EntityTypeReview  = "review"
	EntityTypePayment = "payment"
	EntityTypeComment = "comment"
	EntityTypeRating  = "rating"
	// EntityTypeUser is used by the login/logout activity rows (F001):
	// entity_id is the acting admin's own user id. entity_type has no CHECK
	// constraint, so this needs no migration.
	EntityTypeUser = "user"
)

// ActivityLog represents an audit/activity event in the system.
type ActivityLog struct {
	ID         uuid.UUID       `json:"id"`
	UserID     uuid.UUID       `json:"user_id"`
	User       *User           `json:"user,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   uuid.UUID       `json:"entity_id"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IPAddress  *string         `json:"ip_address,omitempty"`
	UserAgent  *string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
