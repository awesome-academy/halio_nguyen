package domain

import (
	"time"

	"github.com/google/uuid"
)

// Review status constants
const (
	ReviewStatusDraft     = "draft"
	ReviewStatusPublished = "published"
	ReviewStatusHidden    = "hidden"
)

// ReviewCategory represents a review topic (e.g. place, food, news).
type ReviewCategory struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Review represents an article/post about place, food, or news written by users.
type Review struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	User             *User           `json:"user,omitempty"`
	ReviewCategoryID uuid.UUID       `json:"review_category_id"`
	ReviewCategory   *ReviewCategory `json:"review_category,omitempty"`
	Title            string          `json:"title"`
	Slug             string          `json:"slug"`
	Content          string          `json:"content"`
	ThumbnailURL     *string         `json:"thumbnail_url,omitempty"`
	LikeCount        int             `json:"like_count"`
	CommentCount     int             `json:"comment_count"`
	Status           string          `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"`
}

// ReviewLike represents a user liking a review.
type ReviewLike struct {
	UserID    uuid.UUID `json:"user_id"`
	ReviewID  uuid.UUID `json:"review_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Comment represents a comment on a review, supporting recursive replies.
type Comment struct {
	ID        uuid.UUID  `json:"id"`
	ReviewID  uuid.UUID  `json:"review_id"`
	UserID    uuid.UUID  `json:"user_id"`
	User      *User      `json:"user,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Replies   []Comment  `json:"replies,omitempty"`
	Content   string     `json:"content"`
	IsHidden  bool       `json:"is_hidden"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// TourRating represents a 1-5 star score given to a tour by a user with a verified booking.
type TourRating struct {
	ID        uuid.UUID  `json:"id"`
	TourID    uuid.UUID  `json:"tour_id"`
	UserID    uuid.UUID  `json:"user_id"`
	User      *User      `json:"user,omitempty"`
	BookingID uuid.UUID  `json:"booking_id"` // Required: only booked users can rate
	Score     int        `json:"score"`      // 1 to 5
	Comment   *string    `json:"comment,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
