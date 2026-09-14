package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/apperror"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/domain"
	"github.com/sun-booking/sun-booking-tours/apps/api/internal/repository"
)

// TourImageInput is A7's create DTO (also reused by A3's nested images[]).
type TourImageInput struct {
	ImageURL string
	Caption  *string
}

// TourImageUpdateInput is A8's DTO (BR-011: caption and sort_order are
// trusted verbatim, never renumbered).
type TourImageUpdateInput struct {
	Caption   *string
	SortOrder int
}

// TourImageService owns F003's gallery rules (CAP-02): A7 next-position
// assignment, A8's trust-as-is reorder, A9's idempotent remove.
type TourImageService struct {
	db        repository.DB
	tourRepo  repository.TourRepository
	imageRepo repository.TourImageRepository
}

// NewTourImageService wires the service against the pool and its repos.
func NewTourImageService(db repository.DB, tourRepo repository.TourRepository, imageRepo repository.TourImageRepository) *TourImageService {
	return &TourImageService{db: db, tourRepo: tourRepo, imageRepo: imageRepo}
}

// Create is A7: sort_order defaults to the next available position.
func (s *TourImageService) Create(ctx context.Context, tourID uuid.UUID, in TourImageInput) (*domain.TourImage, error) {
	if err := s.requireTour(ctx, tourID); err != nil {
		return nil, err
	}
	img, err := buildTourImage(tourID, in)
	if err != nil {
		return nil, err
	}
	next, err := s.imageRepo.NextSortOrder(ctx, s.db, tourID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	img.SortOrder = next
	created, err := s.imageRepo.Insert(ctx, s.db, img)
	if err != nil {
		return nil, mapTourImageRepoError(err)
	}
	return created, nil
}

// Update is A8 (BR-011 — no renumbering, no gap-filling of sibling rows).
func (s *TourImageService) Update(ctx context.Context, tourID, imageID uuid.UUID, in TourImageUpdateInput) (*domain.TourImage, error) {
	if err := s.requireTour(ctx, tourID); err != nil {
		return nil, err
	}
	updated, err := s.imageRepo.Update(ctx, s.db, tourID, imageID, emptyToNil(in.Caption), in.SortOrder)
	if err != nil {
		return nil, mapTourImageRepoError(err)
	}
	return updated, nil
}

// Delete is A9: deleting an already-deleted (or never-existing) image under
// an existing tour is idempotent — no error surfaces to the admin.
func (s *TourImageService) Delete(ctx context.Context, tourID, imageID uuid.UUID) error {
	if err := s.requireTour(ctx, tourID); err != nil {
		return err
	}
	if err := s.imageRepo.SoftDelete(ctx, s.db, tourID, imageID); err != nil {
		if errors.Is(err, repository.ErrTourImageNotFound) {
			return nil
		}
		return apperror.NewInternal(err)
	}
	return nil
}

func (s *TourImageService) requireTour(ctx context.Context, tourID uuid.UUID) error {
	if _, err := s.tourRepo.FindByID(ctx, s.db, tourID); err != nil {
		return mapTourRepoError(err)
	}
	return nil
}

// buildTourImage validates a TourImageInput (image_url required, http(s)
// only — Security: rendered into <img src>) and shapes it into the row to
// write; sort_order is assigned by the caller (NextSortOrder or array index).
func buildTourImage(tourID uuid.UUID, in TourImageInput) (*domain.TourImage, error) {
	url := strings.TrimSpace(in.ImageURL)
	if url == "" {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", map[string]string{"image_url": "Image URL is required."})
	}
	if !isHTTPURL(url) {
		return nil, apperror.NewUnprocessable("Please correct the highlighted fields.", map[string]string{"image_url": "Please enter a valid http(s) URL."})
	}
	return &domain.TourImage{
		TourID:   tourID,
		ImageURL: url,
		Caption:  emptyToNil(in.Caption),
	}, nil
}

func mapTourImageRepoError(err error) error {
	if errors.Is(err, repository.ErrTourImageNotFound) {
		return apperror.NewNotFound("Tour image not found")
	}
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.NewInternal(err)
}
