package domain

import (
	"time"

	"github.com/google/uuid"
)

// Booking status constants
const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCompleted = "completed"
	BookingStatusCancelled = "cancelled"
)

// Booking represents a tour reservation.
type Booking struct {
	ID                 uuid.UUID     `json:"id"`
	BookingCode        string        `json:"booking_code"`
	UserID             uuid.UUID     `json:"user_id"`
	User               *User         `json:"user,omitempty"`
	TourID             uuid.UUID     `json:"tour_id"`
	Tour               *Tour         `json:"tour,omitempty"`
	ScheduleID         uuid.UUID     `json:"schedule_id"`
	Schedule           *TourSchedule `json:"schedule,omitempty"`
	NumParticipants    int           `json:"num_participants"`
	UnitPrice          float64       `json:"unit_price"`
	TotalPrice         float64       `json:"total_price"`
	ContactName        string        `json:"contact_name"`
	ContactPhone       string        `json:"contact_phone"`
	ContactEmail       string        `json:"contact_email"`
	SpecialRequests    *string       `json:"special_requests,omitempty"`
	Status             string        `json:"status"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CancellationReason *string       `json:"cancellation_reason,omitempty"`
	Payment            *Payment      `json:"payment,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	DeletedAt          *time.Time    `json:"deleted_at,omitempty"`
}
