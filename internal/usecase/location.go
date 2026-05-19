package usecase

import (
	"context"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type LocationUseCase interface {
	Create(ctx context.Context, location *entity.Location) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Location, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Location, error)
	Update(ctx context.Context, location *entity.Location) error
	Delete(ctx context.Context, id int64) error
}

type locationUseCase struct {
	locationRepo repository.LocationRepository
}

func NewLocationUseCase(locationRepo repository.LocationRepository) LocationUseCase {
	return &locationUseCase{
		locationRepo: locationRepo,
	}
}

func (uc *locationUseCase) Create(ctx context.Context, location *entity.Location) (int64, error) {
	return uc.locationRepo.Create(ctx, location)
}

func (uc *locationUseCase) GetByID(ctx context.Context, id int64) (*entity.Location, error) {
	return uc.locationRepo.GetByID(ctx, id)
}

func (uc *locationUseCase) ListByClub(ctx context.Context, clubID int64) ([]entity.Location, error) {
	return uc.locationRepo.ListByClub(ctx, clubID)
}

func (uc *locationUseCase) Update(ctx context.Context, location *entity.Location) error {
	return uc.locationRepo.Update(ctx, location)
}

func (uc *locationUseCase) Delete(ctx context.Context, id int64) error {
	return uc.locationRepo.Delete(ctx, id)
}
