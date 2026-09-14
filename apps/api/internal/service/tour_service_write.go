package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// Create is A3: one transaction across tours -> tour_images -> tour_schedules.
// Any failure (validation, BR-007 category check, a 23505 on either unique
// constraint) rolls back all three tables — no partial tour is ever
// persisted (SC-002).
func (s *TourService) Create(ctx context.Context, in TourInput, images []TourImageInput, schedules []TourScheduleInput) (*domain.Tour, error) {
	tour, err := buildTour(in)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.requireActiveCategory(ctx, tx, in.CategoryID); err != nil {
		return nil, err
	}

	created, err := s.tourRepo.Insert(ctx, tx, tour)
	if err != nil {
		return nil, mapTourRepoError(err)
	}

	for i, imgIn := range images {
		img, err := buildTourImage(created.ID, imgIn)
		if err != nil {
			return nil, err
		}
		img.SortOrder = i
		if _, err := s.imageRepo.Insert(ctx, tx, img); err != nil {
			return nil, mapTourImageRepoError(err)
		}
	}

	for _, schIn := range schedules {
		sch, err := buildTourSchedule(created.ID, schIn)
		if err != nil {
			return nil, err
		}
		if _, err := s.scheduleRepo.Insert(ctx, tx, sch); err != nil {
			return nil, mapTourScheduleRepoError(err)
		}
	}

	return created, commit(ctx, tx)
}

// Update is A4 (BR-001/002/005/007). See TourInput's doc comment on why
// status is out of scope here.
func (s *TourService) Update(ctx context.Context, id uuid.UUID, in TourInput) (*domain.Tour, error) {
	tour, err := buildTour(in)
	if err != nil {
		return nil, err
	}
	tour.ID = id

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	if err := s.requireActiveCategory(ctx, tx, in.CategoryID); err != nil {
		return nil, err
	}
	updated, err := s.tourRepo.Update(ctx, tx, tour)
	if err != nil {
		return nil, mapTourRepoError(err)
	}
	return updated, commit(ctx, tx)
}

// requireActiveCategory is BR-007: category_id must reference an active,
// non-deleted category. Runs against the caller's tx so it observes the
// same snapshot the subsequent write does.
func (s *TourService) requireActiveCategory(ctx context.Context, db repository.DB, categoryID uuid.UUID) error {
	msg := "Selected category is not available."
	cat, err := s.categoryRepo.FindByID(ctx, db, categoryID)
	if errors.Is(err, repository.ErrCategoryNotFound) {
		return apperror.NewUnprocessable(msg, map[string]string{"category_id": msg})
	}
	if err != nil {
		return apperror.NewInternal(err)
	}
	if !cat.IsActive {
		return apperror.NewUnprocessable(msg, map[string]string{"category_id": msg})
	}
	return nil
}
