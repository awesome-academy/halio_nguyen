package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// TourInput is the explicit create/update DTO (never bind into domain.Tour —
// id/slug-derivation/created_at/updated_at/deleted_at are not directly
// bindable). Pointer fields distinguish "omitted" from a zero value, which is
// what makes BR-005 detectable: AvgRating/TotalRatings are never written —
// their only purpose here is to be checked for presence and rejected.
//
// Status is only meaningful on Create: FR-305 makes A5 (UpdateStatus) the
// single writer of tours.status post-creation. This is a judgment call — the
// technical spec's A4 rung says the update body carries "the same base
// fields as A3", which read literally would let A4 silently bypass SM-001.
// BR-009's own Bin-2 listing only names A5 as its user, so status changes
// are kept exclusively behind A5's transition guard; the update repository
// query does not include a status column in its SET clause at all.
type TourInput struct {
	CategoryID      uuid.UUID
	Title           string
	Description     string
	Itinerary       *string
	Destination     string
	DurationDays    int
	DurationNights  int
	Price           float64
	DiscountPrice   *float64
	MaxParticipants int
	ThumbnailURL    *string
	Highlights      []string
	Inclusions      *string
	Exclusions      *string
	Status          *string
	AvgRating       *float64
	TotalRatings    *int
}

// TourScheduleWithPrice is one A2 schedule row plus its ALG-001 effective
// price.
type TourScheduleWithPrice struct {
	domain.TourSchedule
	EffectivePrice float64 `json:"effective_price"`
}

// TourDetail is A2's response shape: the tour composed with its non-deleted
// images and schedules, read from their own repositories rather than one
// triple-join query (phase-04 Implementation Steps #3). Images/Schedules
// here shadow domain.Tour's own (unset) fields of the same JSON name.
type TourDetail struct {
	domain.Tour
	Images    []domain.TourImage      `json:"images"`
	Schedules []TourScheduleWithPrice `json:"schedules"`
}

// TourService owns F003's tour rules: FR-002 slug derivation, BR-001/002/007
// field validation, BR-005 rating-field rejection, BR-009/SM-001 status
// transitions, and BR-012's delete guard.
type TourService struct {
	db           repository.DB
	tourRepo     repository.TourRepository
	categoryRepo repository.CategoryRepository
	imageRepo    repository.TourImageRepository
	scheduleRepo repository.TourScheduleRepository
	bookingRepo  repository.BookingRepository
}

// NewTourService wires the service against the pool (for A3/A6's
// transactions) and its repositories.
func NewTourService(
	db repository.DB,
	tourRepo repository.TourRepository,
	categoryRepo repository.CategoryRepository,
	imageRepo repository.TourImageRepository,
	scheduleRepo repository.TourScheduleRepository,
	bookingRepo repository.BookingRepository,
) *TourService {
	return &TourService{
		db: db, tourRepo: tourRepo, categoryRepo: categoryRepo,
		imageRepo: imageRepo, scheduleRepo: scheduleRepo, bookingRepo: bookingRepo,
	}
}

// List is A1.
func (s *TourService) List(ctx context.Context, p repository.TourListParams) (*repository.Paginated[domain.Tour], error) {
	p.Normalize()
	items, total, err := s.tourRepo.List(ctx, s.db, p)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	return &repository.Paginated[domain.Tour]{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// Get is A2: composes the tour with its images and schedules, attaching
// ALG-001's effective_price to each schedule row.
func (s *TourService) Get(ctx context.Context, id uuid.UUID) (*TourDetail, error) {
	tour, err := s.tourRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, mapTourRepoError(err)
	}
	images, err := s.imageRepo.ListByTour(ctx, s.db, id)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	schedules, err := s.scheduleRepo.ListByTour(ctx, s.db, id)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	withPrice := make([]TourScheduleWithPrice, len(schedules))
	for i, sc := range schedules {
		withPrice[i] = TourScheduleWithPrice{
			TourSchedule:   sc,
			EffectivePrice: EffectivePrice(sc.PriceOverride, tour.DiscountPrice, tour.Price),
		}
	}
	return &TourDetail{Tour: *tour, Images: images, Schedules: withPrice}, nil
}

// UpdateStatus is A5 (BR-009/SM-001). No transaction: a rejected transition
// never calls the repository write, so SC-003 (no DB write on a 422) holds
// without one; the guarded-UPDATE-with-expected-state hardening for this
// race is deferred (decisions.md §7 D-8), not this phase's job.
func (s *TourService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Tour, error) {
	tour, err := s.tourRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, mapTourRepoError(err)
	}
	if !canTransition(tourTransitions, tour.Status, status) {
		msg := "This status change isn't allowed."
		return nil, apperror.NewUnprocessable(msg, map[string]string{"status": msg})
	}
	updated, err := s.tourRepo.UpdateStatus(ctx, s.db, id, status)
	if err != nil {
		return nil, mapTourRepoError(err)
	}
	return updated, nil
}

// Delete is A6 (BR-006/BR-012, D4): lock the row, count active bookings, and
// either refuse with 409 (zero writes) or soft-delete — count and write
// share one transaction so a concurrent booking-create cannot slip between
// them (see Architecture § BR-012 guard).
func (s *TourService) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.tourRepo.LockForUpdate(ctx, tx, id); err != nil {
		return mapTourRepoError(err)
	}
	n, err := s.bookingRepo.CountActiveByTour(ctx, tx, id)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if n > 0 {
		return apperror.NewConflict(
			fmt.Sprintf("Cannot delete: %d active booking(s) reference this tour.", n),
			map[string]string{"booking_count": fmt.Sprintf("%d", n)},
		)
	}
	if err := s.tourRepo.SoftDelete(ctx, tx, id); err != nil {
		return mapTourRepoError(err)
	}
	return commit(ctx, tx)
}

func mapTourRepoError(err error) error {
	if errors.Is(err, repository.ErrTourNotFound) {
		return apperror.NewNotFound("Tour not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
