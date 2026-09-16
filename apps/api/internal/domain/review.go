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

// ReviewListItem is F006 A1's row. Deliberately omits `content` — the list
// renders titles, and shipping every article body on a paginated list is
// waste the client never uses.
type ReviewListItem struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Status           string    `json:"status"`
	LikeCount        int       `json:"like_count"`
	CommentCount     int       `json:"comment_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	UserID           uuid.UUID `json:"user_id"`
	AuthorName       string    `json:"author_name"`
	AuthorAvatarURL  *string   `json:"author_avatar_url,omitempty"`
	ReviewCategoryID uuid.UUID `json:"review_category_id"`
	CategoryName     string    `json:"category_name"`
}

// CommentAuthor is the ONLY author shape F006's comment tree ever ships.
// Never widen it — author identity is joined for display only (Security
// Considerations: PII display-only).
type CommentAuthor struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

// CommentNode is one node of F006 A2's tree. A soft-deleted node is
// REDACTED server-side (comment_thread_builder.go) down to
// {id, parent_id, created_at, is_deleted:true} plus its replies — Content is
// "" and User is nil, so neither ever reaches the wire (R8: shipping content
// alongside a flag and hiding it in the client is a leak, not a redaction).
type CommentNode struct {
	ID        uuid.UUID      `json:"id"`
	ParentID  *uuid.UUID     `json:"parent_id,omitempty"`
	Content   string         `json:"content,omitempty"`
	IsHidden  bool           `json:"is_hidden"`
	IsDeleted bool           `json:"is_deleted"`
	CreatedAt time.Time      `json:"created_at"`
	User      *CommentAuthor `json:"user,omitempty"`
	Replies   []CommentNode  `json:"replies"`
}

// ReviewDetail is F006 A2's response body: the review, its joined
// author/category display fields, and the assembled comment tree.
type ReviewDetail struct {
	*Review
	AuthorName      string        `json:"author_name"`
	AuthorAvatarURL *string       `json:"author_avatar_url,omitempty"`
	CategoryName    string        `json:"category_name"`
	Comments        []CommentNode `json:"comments"`
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
