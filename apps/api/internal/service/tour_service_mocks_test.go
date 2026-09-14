package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// mustDate parses a YYYY-MM-DD literal, panicking on a malformed test fixture
// (shared by tour_service_test.go and tour_schedule_service_test.go).
func mustDate(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

// --- mocks (reused by tour_image_service_test.go / tour_schedule_service_test.go) ---
// mockTourScheduleRepo, mockBookingRepo, and the shared test helpers live in
// tour_service_helpers_test.go to keep both files under the 200-line cap.

type mockTourRepo struct {
	repository.TourRepository
	listItems       []domain.Tour
	listTotal       int64
	found           *domain.Tour
	findErr         error
	inserted        *domain.Tour
	insertErr       error
	updateErr       error
	updateStatusErr error
	softDeleteCalls int
	softDeleteErr   error
	lockErr         error
}

func (m *mockTourRepo) List(context.Context, repository.DB, repository.TourListParams) ([]domain.Tour, int64, error) {
	return m.listItems, m.listTotal, nil
}
func (m *mockTourRepo) FindByID(context.Context, repository.DB, uuid.UUID) (*domain.Tour, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.found, nil
}
func (m *mockTourRepo) Insert(_ context.Context, _ repository.DB, t *domain.Tour) (*domain.Tour, error) {
	m.inserted = t
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	t.ID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	return t, nil
}
func (m *mockTourRepo) Update(_ context.Context, _ repository.DB, t *domain.Tour) (*domain.Tour, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return t, nil
}
func (m *mockTourRepo) UpdateStatus(_ context.Context, _ repository.DB, id uuid.UUID, status string) (*domain.Tour, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	updated := *m.found
	updated.Status = status
	return &updated, nil
}
func (m *mockTourRepo) SoftDelete(context.Context, repository.DB, uuid.UUID) error {
	m.softDeleteCalls++
	return m.softDeleteErr
}
func (m *mockTourRepo) LockForUpdate(context.Context, repository.DB, uuid.UUID) error {
	return m.lockErr
}

type mockTourCategoryRepo struct {
	repository.CategoryRepository
	category *domain.Category
	err      error
}

func (m *mockTourCategoryRepo) FindByID(context.Context, repository.DB, uuid.UUID) (*domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.category, nil
}

type mockTourImageRepo struct {
	repository.TourImageRepository
	inserted        []*domain.TourImage
	insertFailOn    int // 1-indexed call number that fails; 0 = never
	insertCalls     int
	nextSortOrder   int
	listItems       []domain.TourImage
	updated         *domain.TourImage
	updateErr       error
	softDeleteCalls int
	softDeleteErr   error
}

func (m *mockTourImageRepo) Insert(_ context.Context, _ repository.DB, img *domain.TourImage) (*domain.TourImage, error) {
	m.insertCalls++
	if m.insertFailOn != 0 && m.insertCalls == m.insertFailOn {
		return nil, repository.TourConflict("image_url")
	}
	m.inserted = append(m.inserted, img)
	return img, nil
}
func (m *mockTourImageRepo) NextSortOrder(context.Context, repository.DB, uuid.UUID) (int, error) {
	return m.nextSortOrder, nil
}
func (m *mockTourImageRepo) ListByTour(context.Context, repository.DB, uuid.UUID) ([]domain.TourImage, error) {
	return m.listItems, nil
}
func (m *mockTourImageRepo) Update(_ context.Context, _ repository.DB, tourID, id uuid.UUID, caption *string, sortOrder int) (*domain.TourImage, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	img := &domain.TourImage{ID: id, TourID: tourID, Caption: caption, SortOrder: sortOrder}
	m.updated = img
	return img, nil
}
func (m *mockTourImageRepo) SoftDelete(context.Context, repository.DB, uuid.UUID, uuid.UUID) error {
	m.softDeleteCalls++
	return m.softDeleteErr
}
