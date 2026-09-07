package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserEntityInitialization(t *testing.T) {
	uid := uuid.New()
	email := "traveler@sunbooking.com"
	name := "Alice Traveler"
	now := time.Now()

	user := User{
		ID:        uid,
		Email:     email,
		FullName:  name,
		Role:      RoleUser,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if user.ID != uid {
		t.Errorf("expected ID %v, got %v", uid, user.ID)
	}
	if user.Role != RoleUser {
		t.Errorf("expected role '%s', got '%s'", RoleUser, user.Role)
	}
	if !user.IsActive {
		t.Errorf("expected user to be active")
	}
	if user.DeletedAt != nil {
		t.Errorf("expected DeletedAt to be nil on new user")
	}
}

func TestRoleConstants(t *testing.T) {
	if RoleAdmin != "admin" {
		t.Errorf("expected RoleAdmin to be 'admin', got '%s'", RoleAdmin)
	}
	if RoleUser != "user" {
		t.Errorf("expected RoleUser to be 'user', got '%s'", RoleUser)
	}
}

func TestActivityActionConstants(t *testing.T) {
	expectedActions := map[string]string{
		"booking_tour":  ActionBookingTour,
		"cancel_tour":   ActionCancelTour,
		"create_review": ActionCreateReview,
		"pay_tour":      ActionPayTour,
	}

	for key, val := range expectedActions {
		if key != val {
			t.Errorf("expected action '%s', got '%s'", key, val)
		}
	}
}
