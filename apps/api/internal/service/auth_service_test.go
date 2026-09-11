package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

const testJWTSecret = "test-secret-at-least-32-bytes-long!"

// activeAdminPassword is the plaintext behind adminBcryptHash below.
const activeAdminPassword = "Correct@123456"

// adminBcryptHash is a real bcrypt cost-10 hash of activeAdminPassword,
// generated once and pinned here — the mock never calls bcrypt itself.
const adminBcryptHash = "$2a$10$..QrZV9la86HLhpx6rA/W.qDWZXcK2mB/CyFMa0SrejjnSgZO4DAu"

type stubUserRepo struct {
	user *domain.User
	err  error
}

func (s stubUserRepo) FindActiveAdminByEmail(ctx context.Context, db repository.DB, email string) (*domain.User, error) {
	return s.user, s.err
}

func (s stubUserRepo) FindByID(ctx context.Context, db repository.DB, id uuid.UUID) (*domain.User, error) {
	return s.user, s.err
}

type stubActivityRepo struct {
	inserts int
}

func (s *stubActivityRepo) Insert(ctx context.Context, db repository.DB, log repository.NewActivityLog) error {
	s.inserts++
	return nil
}

func hashedUser(role string, isActive bool) *domain.User {
	hash := adminBcryptHash
	return &domain.User{
		ID:           uuid.New(),
		Email:        "admin@sunbooking.com",
		PasswordHash: &hash,
		FullName:     "System Administrator",
		Role:         role,
		IsActive:     isActive,
	}
}

func newTestService(users repository.UserRepository, activity repository.ActivityLogRepository) *AuthService {
	throttle := NewEmailThrottle(1000, time.Minute)
	return NewAuthService(nil, users, activity, throttle, testJWTSecret, time.Hour)
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc := newTestService(stubUserRepo{user: hashedUser(domain.RoleAdmin, true)}, &stubActivityRepo{})
	_, err := svc.Login(context.Background(), LoginInput{Email: "admin@sunbooking.com", Password: "wrong-password"})
	assertGenericUnauthorized(t, err)
}

func TestLoginRejectsNonAdminRole(t *testing.T) {
	svc := newTestService(stubUserRepo{user: hashedUser(domain.RoleUser, true)}, &stubActivityRepo{})
	_, err := svc.Login(context.Background(), LoginInput{Email: "user@sunbooking.com", Password: activeAdminPassword})
	assertGenericUnauthorized(t, err)
}

func TestLoginRejectsInactiveAccount(t *testing.T) {
	svc := newTestService(stubUserRepo{user: hashedUser(domain.RoleAdmin, false)}, &stubActivityRepo{})
	_, err := svc.Login(context.Background(), LoginInput{Email: "admin@sunbooking.com", Password: activeAdminPassword})
	assertGenericUnauthorized(t, err)
}

func TestLoginRejectsUnknownEmailSameAsSoftDeleted(t *testing.T) {
	// FindActiveAdminByEmail already filters deleted_at IS NULL in SQL, so a
	// soft-deleted account surfaces to the service exactly like an unknown
	// email: repository.ErrUserNotFound.
	svc := newTestService(stubUserRepo{err: repository.ErrUserNotFound}, &stubActivityRepo{})
	_, err := svc.Login(context.Background(), LoginInput{Email: "ghost@sunbooking.com", Password: activeAdminPassword})
	assertGenericUnauthorized(t, err)
}

func TestLoginRejectionMessagesAreIdentical(t *testing.T) {
	cases := map[string]*AuthService{
		"wrong password":     newTestService(stubUserRepo{user: hashedUser(domain.RoleAdmin, true)}, &stubActivityRepo{}),
		"non-admin role":     newTestService(stubUserRepo{user: hashedUser(domain.RoleUser, true)}, &stubActivityRepo{}),
		"inactive account":   newTestService(stubUserRepo{user: hashedUser(domain.RoleAdmin, false)}, &stubActivityRepo{}),
		"unknown/deleted":    newTestService(stubUserRepo{err: repository.ErrUserNotFound}, &stubActivityRepo{}),
	}

	var messages []string
	for _, svc := range cases {
		_, err := svc.Login(context.Background(), LoginInput{Email: "x@x.com", Password: "wrong"})
		appErr := err.(*apperror.Error)
		messages = append(messages, appErr.Message)
	}
	for _, m := range messages {
		if m != messages[0] {
			t.Fatalf("expected every rejection path to share one message, got %v", messages)
		}
	}
}

func TestLoginHappyPathIssuesOneHourToken(t *testing.T) {
	activity := &stubActivityRepo{}
	svc := newTestService(stubUserRepo{user: hashedUser(domain.RoleAdmin, true)}, activity)

	result, err := svc.Login(context.Background(), LoginInput{Email: "admin@sunbooking.com", Password: activeAdminPassword, From: "/admin/tours"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected a non-empty token")
	}
	if result.RedirectTo != "/admin/tours" {
		t.Errorf("RedirectTo = %q, want %q", result.RedirectTo, "/admin/tours")
	}
	if activity.inserts != 1 {
		t.Errorf("expected exactly one activity log insert, got %d", activity.inserts)
	}
}

func TestLoginTripsRateLimit(t *testing.T) {
	throttle := NewEmailThrottle(1, time.Minute)
	defer throttle.Close()
	svc := NewAuthService(nil, stubUserRepo{user: hashedUser(domain.RoleAdmin, true)}, &stubActivityRepo{}, throttle, testJWTSecret, time.Hour)

	if _, err := svc.Login(context.Background(), LoginInput{Email: "admin@sunbooking.com", Password: "wrong"}); err == nil {
		t.Fatal("expected first attempt to be processed (and rejected on password)")
	}
	_, err := svc.Login(context.Background(), LoginInput{Email: "admin@sunbooking.com", Password: activeAdminPassword})
	appErr, ok := err.(*apperror.Error)
	if !ok || appErr.Code != apperror.CodeRateLimited {
		t.Fatalf("expected the 2nd attempt to be rate limited, got %v", err)
	}
}

func assertGenericUnauthorized(t *testing.T, err error) {
	t.Helper()
	appErr, ok := err.(*apperror.Error)
	if !ok {
		t.Fatalf("expected *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("expected unauthorized, got %v", appErr.Code)
	}
	if appErr.Message != invalidCredentialsMessage {
		t.Errorf("expected generic message %q, got %q", invalidCredentialsMessage, appErr.Message)
	}
}
