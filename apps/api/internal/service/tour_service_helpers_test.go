package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// mockTourScheduleRepo, mockBookingRepo, and the fixture/assertion helpers
// below are shared by tour_service_test.go, tour_image_service_test.go and
// tour_schedule_service_test.go; the repo-level mocks live in
// tour_service_mocks_test.go (split to keep both files under the 200-line cap).

type mockTourScheduleRepo struct {
	repository.TourScheduleRepository
	inserted        []*domain.TourSchedule
	insertFailOn    int
	insertCalls     int
	listItems       []domain.TourSchedule
	found           *domain.TourSchedule
	findErr         error
	updateErr       error
	updateStatusErr error
	softDeleteCalls int
	softDeleteErr   error
	lockErr         error
}

func (m *mockTourScheduleRepo) Insert(_ context.Context, _ repository.DB, s *domain.TourSchedule) (*domain.TourSchedule, error) {
	m.insertCalls++
	if m.insertFailOn != 0 && m.insertCalls == m.insertFailOn {
		return nil, repository.ScheduleDateConflict("departure_date")
	}
	m.inserted = append(m.inserted, s)
	return s, nil
}
func (m *mockTourScheduleRepo) ListByTour(context.Context, repository.DB, uuid.UUID) ([]domain.TourSchedule, error) {
	return m.listItems, nil
}
func (m *mockTourScheduleRepo) FindByID(context.Context, repository.DB, uuid.UUID, uuid.UUID) (*domain.TourSchedule, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.found, nil
}
func (m *mockTourScheduleRepo) Update(_ context.Context, _ repository.DB, s *domain.TourSchedule) (*domain.TourSchedule, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return s, nil
}
func (m *mockTourScheduleRepo) UpdateStatus(_ context.Context, _ repository.DB, tourID, id uuid.UUID, status string) (*domain.TourSchedule, error) {
	if m.updateStatusErr != nil {
		return nil, m.updateStatusErr
	}
	updated := *m.found
	updated.Status = status
	return &updated, nil
}
func (m *mockTourScheduleRepo) SoftDelete(context.Context, repository.DB, uuid.UUID, uuid.UUID) error {
	m.softDeleteCalls++
	return m.softDeleteErr
}
func (m *mockTourScheduleRepo) LockForUpdate(context.Context, repository.DB, uuid.UUID, uuid.UUID) error {
	return m.lockErr
}

type mockBookingRepo struct {
	repository.BookingRepository
	tourCount     int64
	scheduleCount int64
}

func (m *mockBookingRepo) CountActiveByTour(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return m.tourCount, nil
}
func (m *mockBookingRepo) CountActiveBySchedule(context.Context, repository.DB, uuid.UUID) (int64, error) {
	return m.scheduleCount, nil
}

// --- helpers ---

func activeCategory() *mockTourCategoryRepo {
	return &mockTourCategoryRepo{category: &domain.Category{ID: uuid.New(), IsActive: true}}
}

func validTourInput() TourInput {
	return TourInput{
		CategoryID: uuid.New(), Title: "Ha Long Bay Cruise", Description: "A lovely cruise",
		Destination: "Ha Long", DurationDays: 3, DurationNights: 2, Price: 100, MaxParticipants: 20,
	}
}

func newTourSvc(tourRepo *mockTourRepo, catRepo *mockTourCategoryRepo, imgRepo *mockTourImageRepo, schRepo *mockTourScheduleRepo, bookingRepo *mockBookingRepo) (*TourService, *fakeTx) {
	tx := &fakeTx{}
	return NewTourService(&fakeDB{tx: tx}, tourRepo, catRepo, imgRepo, schRepo, bookingRepo), tx
}

func appErrCode(t *testing.T, err error) apperror.Code {
	t.Helper()
	appErr, ok := err.(*apperror.Error)
	if !ok {
		t.Fatalf("want *apperror.Error, got %T (%v)", err, err)
	}
	return appErr.Code
}
