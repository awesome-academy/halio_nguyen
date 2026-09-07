package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tour status constants
const (
	TourStatusDraft     = "draft"
	TourStatusPublished = "published"
	TourStatusArchived  = "archived"
)

// Schedule status constants
const (
	ScheduleStatusOpen      = "open"
	ScheduleStatusClosed    = "closed"
	ScheduleStatusCancelled = "cancelled"
)

// Category represents a tour category (e.g. Island, Mountain, Cultural).
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	ImageURL    *string    `json:"image_url,omitempty"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Tour represents a travel tour package.
type Tour struct {
	ID              uuid.UUID       `json:"id"`
	CategoryID      uuid.UUID       `json:"category_id"`
	Category        *Category       `json:"category,omitempty"`
	Title           string          `json:"title"`
	Slug            string          `json:"slug"`
	Description     string          `json:"description"`
	Itinerary       *string         `json:"itinerary,omitempty"`
	Destination     string          `json:"destination"`
	DurationDays    int             `json:"duration_days"`
	DurationNights  int             `json:"duration_nights"`
	Price           float64         `json:"price"`
	DiscountPrice   *float64        `json:"discount_price,omitempty"`
	MaxParticipants int             `json:"max_participants"`
	ThumbnailURL    *string         `json:"thumbnail_url,omitempty"`
	Highlights      []string        `json:"highlights,omitempty"`
	Inclusions      *string         `json:"inclusions,omitempty"`
	Exclusions      *string         `json:"exclusions,omitempty"`
	Status          string          `json:"status"`
	AvgRating       float64         `json:"avg_rating"`
	TotalRatings    int             `json:"total_ratings"`
	Images          []TourImage     `json:"images,omitempty"`
	Schedules       []TourSchedule  `json:"schedules,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty"`
}

// TourImage represents gallery photos for a tour.
type TourImage struct {
	ID        uuid.UUID  `json:"id"`
	TourID    uuid.UUID  `json:"tour_id"`
	ImageURL  string     `json:"image_url"`
	Caption   *string    `json:"caption,omitempty"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// TourSchedule represents departure departures and slot availability.
type TourSchedule struct {
	ID             uuid.UUID  `json:"id"`
	TourID         uuid.UUID  `json:"tour_id"`
	DepartureDate  time.Time  `json:"departure_date"`
	ReturnDate     time.Time  `json:"return_date"`
	AvailableSlots int        `json:"available_slots"`
	PriceOverride  *float64   `json:"price_override,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}
