package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role constants
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// OAuthProvider constants
const (
	OAuthProviderGoogle   = "google"
	OAuthProviderFacebook = "facebook"
	OAuthProviderTwitter  = "twitter"
)

// User represents a registered user or administrator.
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash *string    `json:"-"`
	FullName     string     `json:"full_name"`
	Phone        *string    `json:"phone,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// UserDetail is A2's response shape (phase-07): the full profile plus its
// two LIFETIME history counts. Deliberately embeds *User rather than being a
// hand-built map — User.PasswordHash's json:"-" is what keeps the hash off
// the wire, and that guarantee only holds if this type reuses User instead
// of re-declaring its fields. The embedding is anonymous so encoding/json
// promotes the profile fields to the top level: A2's body is one flat object
// (FR-003's "full profile plus counts"), not a nested {"user": {...}} wrapper.
type UserDetail struct {
	*User
	BookingCount int64 `json:"booking_count"`
	ReviewCount  int64 `json:"review_count"`
}

// UserOAuthAccount represents a connected social authentication account.
type UserOAuthAccount struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	Provider       string     `json:"provider"`
	ProviderUserID string     `json:"provider_user_id"`
	Email          *string    `json:"email,omitempty"`
	AccessToken    *string    `json:"-"`
	RefreshToken   *string    `json:"-"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// UserBankAccount represents a user bank account for internet banking.
type UserBankAccount struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	BankName          string     `json:"bank_name"`
	BankCode          *string    `json:"bank_code,omitempty"`
	AccountNumber     string     `json:"account_number"`
	AccountHolderName string     `json:"account_holder_name"`
	IsDefault         bool       `json:"is_default"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}
