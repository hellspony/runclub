package usecase

import (
	"context"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type RaceUseCase interface {
	Create(ctx context.Context, race *entity.Race) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Race, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Race, error)
	Update(ctx context.Context, race *entity.Race) error
	Delete(ctx context.Context, id int64) error
	RegisterMember(ctx context.Context, raceID, memberID int64, distance string) error
	UnregisterMember(ctx context.Context, raceID, memberID int64) error
	ListRegistrations(ctx context.Context, raceID int64) ([]entity.RaceRegistration, error)
}

type raceUseCase struct {
	raceRepo    repository.RaceRepository
	raceRegRepo repository.RaceRegistrationRepository
}

func NewRaceUseCase(
	raceRepo repository.RaceRepository,
	raceRegRepo repository.RaceRegistrationRepository,
) RaceUseCase {
	return &raceUseCase{
		raceRepo:    raceRepo,
		raceRegRepo: raceRegRepo,
	}
}

func (uc *raceUseCase) Create(ctx context.Context, race *entity.Race) (int64, error) {
	return uc.raceRepo.Create(ctx, race)
}

func (uc *raceUseCase) GetByID(ctx context.Context, id int64) (*entity.Race, error) {
	return uc.raceRepo.GetByID(ctx, id)
}

func (uc *raceUseCase) ListByClub(ctx context.Context, clubID int64) ([]entity.Race, error) {
	return uc.raceRepo.ListByClub(ctx, clubID)
}

func (uc *raceUseCase) Update(ctx context.Context, race *entity.Race) error {
	return uc.raceRepo.Update(ctx, race)
}

func (uc *raceUseCase) Delete(ctx context.Context, id int64) error {
	return uc.raceRepo.Delete(ctx, id)
}

func (uc *raceUseCase) RegisterMember(ctx context.Context, raceID, memberID int64, distance string) error {
	reg := &entity.RaceRegistration{
		RaceID:   raceID,
		MemberID: memberID,
		Distance: distance,
	}

	_, err := uc.raceRegRepo.Create(ctx, reg)
	return err
}

func (uc *raceUseCase) UnregisterMember(ctx context.Context, raceID, memberID int64) error {
	reg, err := uc.raceRegRepo.GetByRaceAndMember(ctx, raceID, memberID)
	if err != nil {
		return err
	}

	return uc.raceRegRepo.Delete(ctx, reg.ID)
}

func (uc *raceUseCase) ListRegistrations(ctx context.Context, raceID int64) ([]entity.RaceRegistration, error) {
	return uc.raceRegRepo.ListByRace(ctx, raceID)
}
