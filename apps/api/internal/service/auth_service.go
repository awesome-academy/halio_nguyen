package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// dummyBcryptHash is a real cost-10 bcrypt hash of a value nobody will ever
// submit as a password. Login always compares against a real hash — this
// one when the looked-up email doesn't exist — so a non-existent email
// costs the same wall-clock time as a wrong password (BR-001 anti-
// enumeration, R4).
const dummyBcryptHash = "$2a$10$ZUWCjBrvW9sUo5dxwmSXGOro0PHNQXXzKTx/PfbtDD.W0czJfETGi"

// invalidCredentialsMessage is returned for every login rejection reason —
// wrong password, non-admin role, inactive, soft-deleted, unknown email —
// so the response never discloses which check failed.
const invalidCredentialsMessage = "Invalid email or password"

// AuthService implements F001: issuing, verifying, and ending the admin
// session. It owns BR-001 (anti-enumeration login gate) and DEC-001
// (redirect validation).
type AuthService struct {
	db       repository.DB
	users    repository.UserRepository
	activity repository.ActivityLogRepository
	throttle *EmailThrottle

	jwtSecret string
	tokenTTL  time.Duration
}

// NewAuthService wires an AuthService against the given repositories, the
// shared DB handle, the per-email throttle (D3), the JWT signing secret,
// and the access-token TTL (D2 — 1h from config).
func NewAuthService(
	db repository.DB,
	users repository.UserRepository,
	activity repository.ActivityLogRepository,
	throttle *EmailThrottle,
	jwtSecret string,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		db:        db,
		users:     users,
		activity:  activity,
		throttle:  throttle,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

// LoginInput is AuthService.Login's request.
type LoginInput struct {
	Email     string
	Password  string
	From      string
	IP        string
	UserAgent string
}

// LoginResult is AuthService.Login's response: the profile to return to the
// client, the signed session token to set as the sun_admin_token cookie,
// and the validated redirect target (DEC-001).
type LoginResult struct {
	User       *domain.User
	Token      string
	RedirectTo string
}

// Login validates credentials and, on success, issues a 1h admin session.
// Order matters (step 7 of phase-02): throttle first, then the lookup, then
// bcrypt is ALWAYS run — against a fixed dummy hash when the email doesn't
// exist — so every rejection path costs the same wall-clock time before the
// role/active check runs and a single generic error is returned.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	if !s.throttle.Allow(in.Email) {
		return nil, apperror.NewRateLimited("Too many login attempts. Please try again later.")
	}

	user, lookupErr := s.users.FindActiveAdminByEmail(ctx, s.db, in.Email)
	if lookupErr != nil && !errors.Is(lookupErr, repository.ErrUserNotFound) {
		return nil, apperror.NewInternal(lookupErr)
	}

	hash := dummyBcryptHash
	if user != nil && user.PasswordHash != nil {
		hash = *user.PasswordHash
	}
	pwErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password))

	isAdminActive := user != nil && user.Role == domain.RoleAdmin && user.IsActive
	if pwErr != nil || !isAdminActive {
		logRejectedLogin(in.Email, in.IP)
		return nil, apperror.NewUnauthorized(invalidCredentialsMessage)
	}

	token, err := jwtutil.Issue(s.jwtSecret, s.tokenTTL, user.ID.String(), user.Role)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	s.logActivityBestEffort(ctx, user.ID, domain.ActionLogin, in.IP, in.UserAgent)

	return &LoginResult{
		User:       user,
		Token:      token,
		RedirectTo: ValidateRedirect(in.From),
	}, nil
}

// Logout records a best-effort activity_logs row when userID is non-nil (a
// valid token happened to be present); it never fails the request — the
// cookie clear is unconditional and lives in the handler (BR-002).
func (s *AuthService) Logout(ctx context.Context, userID *uuid.UUID, ip, userAgent string) {
	if userID == nil {
		return
	}
	s.logActivityBestEffort(ctx, *userID, domain.ActionLogout, ip, userAgent)
}

// Me loads the current admin's profile for GET /auth/me. userID comes from
// the JWT claim the A0 gate already verified.
func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.NewUnauthorized(invalidCredentialsMessage)
	}

	user, err := s.users.FindByID(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, apperror.NewUnauthorized(invalidCredentialsMessage)
		}
		return nil, apperror.NewInternal(err)
	}
	return user, nil
}

