package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/jwtutil"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/service"
)

// F006's A0 gate plus the scaffolding comment_admin_handler_test.go shares.
// Split to keep each file under the 200-line cap — same arrangement as the
// user handler tests.

const (
	testReviewAdminActorID = "88888888-8888-8888-8888-888888888001"
	testReviewID           = "88888888-8888-8888-8888-888888888002"
	testCommentID          = "88888888-8888-8888-8888-888888888003"
)

// stubReviewAdminRepo answers every ReviewAdminRepository method with
// fixtures — these tests exercise routing/parsing/gating, not repository
// behaviour (covered in the repository and service packages).
type stubReviewAdminRepo struct {
	repository.ReviewAdminRepository
	listItems  []domain.ReviewListItem
	getResult  *repository.ReviewWithAuthor
	getErr     error
	updated    *domain.Review
	updateErr  error
	deleteErr  error
	categories []domain.ReviewCategory
}

func (s *stubReviewAdminRepo) List(context.Context, repository.DB, repository.ReviewListParams) ([]domain.ReviewListItem, int64, error) {
	return s.listItems, int64(len(s.listItems)), nil
}

func (s *stubReviewAdminRepo) Get(context.Context, repository.DB, uuid.UUID) (*repository.ReviewWithAuthor, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.getResult != nil {
		return s.getResult, nil
	}
	return &repository.ReviewWithAuthor{Review: domain.Review{ID: uuid.MustParse(testReviewID)}}, nil
}

func (s *stubReviewAdminRepo) ListCommentsForReview(context.Context, repository.DB, uuid.UUID) ([]domain.Comment, error) {
	return nil, nil
}

func (s *stubReviewAdminRepo) UpdateStatus(_ context.Context, _ repository.DB, id uuid.UUID, status string) (*domain.Review, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	if s.updated != nil {
		return s.updated, nil
	}
	return &domain.Review{ID: id, Status: status}, nil
}

func (s *stubReviewAdminRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error {
	return s.deleteErr
}

func (s *stubReviewAdminRepo) ListCategories(context.Context, repository.DB) ([]domain.ReviewCategory, error) {
	return s.categories, nil
}

// stubCommentAdminRepo answers every CommentAdminRepository method.
// mismatch models a comment id paired with the wrong review id (R6/IDOR): a
// real repository's `AND review_id = @review_id` predicate returns zero
// rows for that pairing, which this stub reproduces via ErrCommentNotFound.
type stubCommentAdminRepo struct {
	repository.CommentAdminRepository
	updateResult *domain.Comment
	mismatch     bool
}

func (s *stubCommentAdminRepo) UpdateVisibility(_ context.Context, _ repository.DB, reviewID, commentID uuid.UUID, isHidden bool) (*domain.Comment, error) {
	if s.mismatch {
		return nil, repository.ErrCommentNotFound
	}
	if s.updateResult != nil {
		return s.updateResult, nil
	}
	return &domain.Comment{ID: commentID, ReviewID: reviewID, IsHidden: isHidden}, nil
}

func (s *stubCommentAdminRepo) SoftDelete(context.Context, repository.DB, uuid.UUID, uuid.UUID) error {
	if s.mismatch {
		return repository.ErrCommentNotFound
	}
	return nil
}

// reviewAdminGate wires both handlers behind the same A0 chain router.New
// installs, so the gate is proven by request rather than by inspection.
func reviewAdminGate(t *testing.T, reviewRepo repository.ReviewAdminRepository, commentRepo repository.CommentAdminRepository) *echo.Echo {
	t.Helper()
	e := echo.New()
	e.HTTPErrorHandler = apperror.Handler
	gated := e.Group("/api/v1/admin")
	gated.Use(jwtGate(t))

	reviewSvc := service.NewReviewModerationService(stubUserAdminDB{}, reviewRepo)
	commentSvc := service.NewCommentModerationService(stubUserAdminDB{}, commentRepo)
	NewReviewAdminHandler(reviewSvc, commentSvc).RegisterRoutes(gated)
	return e
}

func reviewAdminRequest(t *testing.T, method, path, body, subject, role string) *http.Request {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, time.Hour, subject, role)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "sun_admin_token", Value: token})
	return req
}

// reviewAdminRoutes lists all 7 F006 actions against fixed ids.
func reviewAdminRoutes(reviewID, commentID string) []struct{ method, path string } {
	return []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/reviews"},
		{http.MethodGet, "/api/v1/admin/reviews/" + reviewID},
		{http.MethodPatch, "/api/v1/admin/reviews/" + reviewID + "/status"},
		{http.MethodDelete, "/api/v1/admin/reviews/" + reviewID},
		{http.MethodPatch, "/api/v1/admin/reviews/" + reviewID + "/comments/" + commentID + "/visibility"},
		{http.MethodDelete, "/api/v1/admin/reviews/" + reviewID + "/comments/" + commentID},
		{http.MethodGet, "/api/v1/admin/review-categories"},
	}
}

func TestAllReviewAdminRoutesRejectMissingCookie(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	for _, r := range reviewAdminRoutes(testReviewID, testCommentID) {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without cookie: %d, want 401", r.method, r.path, rec.Code)
		}
	}
}

func TestAllReviewAdminRoutesRejectNonAdminRole(t *testing.T) {
	e := reviewAdminGate(t, &stubReviewAdminRepo{}, &stubCommentAdminRepo{})
	for _, r := range reviewAdminRoutes(testReviewID, testCommentID) {
		req := reviewAdminRequest(t, r.method, r.path, "", testReviewAdminActorID, "user")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with role=user: %d, want 403", r.method, r.path, rec.Code)
		}
	}
}
