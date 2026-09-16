package domain

import (
	"encoding/json"
	"strings"
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

// TestBankAccountNumberIsNeverSerialized locks in the Phase 10 hardening of
// Phase 6's L1: today the only thing keeping an account number out of a
// response is that no query populates UserBankAccount. If someone adds that
// join, this tag is what stops the number reaching a client.
func TestBankAccountNumberIsNeverSerialized(t *testing.T) {
	b, err := json.Marshal(UserBankAccount{
		AccountNumber:     "1234567890123456",
		AccountHolderName: "Nguyen Van A",
		BankName:          "Vietcombank",
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(b), "1234567890123456") || strings.Contains(string(b), "account_number") {
		t.Errorf("account number leaked into JSON: %s", b)
	}
	// The rest of the struct must still serialize — this is a redaction, not
	// a blanket opt-out.
	if !strings.Contains(string(b), "Vietcombank") {
		t.Errorf("bank_name should still serialize: %s", b)
	}
}
