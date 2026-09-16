package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Mocks and fixtures shared by booking_service_test.go (transitions) and
// booking_cancel_service_test.go (A4), split out to keep every file under the
// 200-line cap — same arrangement as tour_service_mocks_test.go.

// bookingGuardRepo models the one behaviour that matters: the guarded UPDATE
// matches only when the booking's current status is in `expected`. A
// non-match returns ErrBookingNotFound, exactly as the real repository does
// on a zero-row RETURNING, so the 404-vs-409 split is exercised for real
// rather than stubbed into always-succeeding.
type bookingGuardRepo struct {
	repository.BookingRepository
	current     string
	missing     bool
	guardCalls  int
	cancelCalls int
	lastNext    string
	lastReason  string
}

func (m *bookingGuardRepo) row(next string) *domain.Booking {
	return &domain.Booking{
		ID: testBookingID, ScheduleID: testScheduleID, NumParticipants: 2, Status: next,
	}
}

func (m *bookingGuardRepo) UpdateStatusGuarded(_ context.Context, _ repository.DB, _ uuid.UUID, expected []string, next string) (*domain.Booking, error) {
	m.guardCalls++
	if m.missing || !contains(expected, m.current) {
		return nil, repository.ErrBookingNotFound
	}
	m.lastNext = next
	return m.row(next), nil
}

func (m *bookingGuardRepo) CancelGuarded(_ context.Context, _ repository.DB, _ uuid.UUID, reason string) (*domain.Booking, error) {
	m.cancelCalls++
	expected := []string{domain.BookingStatusPending, domain.BookingStatusConfirmed}
	if m.missing || !contains(expected, m.current) {
		return nil, repository.ErrBookingNotFound
	}
	m.lastReason = reason
	return m.row(domain.BookingStatusCancelled), nil
}

func (m *bookingGuardRepo) CurrentStatus(context.Context, repository.DB, uuid.UUID) (string, error) {
	if m.missing {
		return "", repository.ErrBookingNotFound
	}
	return m.current, nil
}

type slotRestoreRepo struct {
	repository.TourScheduleRepository
	restoredBy  int
	restoreCall int
	err         error
}

func (m *slotRestoreRepo) RestoreSlots(_ context.Context, _ repository.DB, _ uuid.UUID, n int) error {
	m.restoreCall++
	m.restoredBy += n
	return m.err
}

type recordingActivityRepo struct {
	logs []repository.NewActivityLog
	err  error
}

func (m *recordingActivityRepo) Insert(_ context.Context, _ repository.DB, log repository.NewActivityLog) error {
	if m.err != nil {
		return m.err
	}
	m.logs = append(m.logs, log)
	return nil
}

var (
	testBookingID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testScheduleID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testAdminID    = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func newBookingSvc(status string) (*BookingService, *bookingGuardRepo, *slotRestoreRepo, *recordingActivityRepo, *fakeTx) {
	bookings := &bookingGuardRepo{current: status}
	schedules := &slotRestoreRepo{}
	activity := &recordingActivityRepo{}
	tx := &fakeTx{}
	return NewBookingService(&fakeDB{tx: tx}, bookings, schedules, activity), bookings, schedules, activity, tx
}

func cancelInput(reason string) CancelInput {
	return CancelInput{Reason: reason, Actor: testAdminID, IPAddress: "203.0.113.7", UserAgent: "test-agent"}
}

func confirmAction(s *BookingService) error {
	_, err := s.Confirm(context.Background(), testBookingID, testAdminID)
	return err
}

func completeAction(s *BookingService) error {
	_, err := s.Complete(context.Background(), testBookingID, testAdminID)
	return err
}

func cancelAction(s *BookingService) error {
	_, err := s.Cancel(context.Background(), testBookingID, cancelInput("Customer changed plans"))
	return err
}
