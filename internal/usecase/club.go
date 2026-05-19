package usecase

import (
	"context"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type ClubUseCase interface {
	Create(ctx context.Context, club *entity.Club) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Club, error)
	GetByTelegramChatID(ctx context.Context, chatID int64) (*entity.Club, error)
	List(ctx context.Context) ([]entity.Club, error)
	Update(ctx context.Context, club *entity.Club) error
	Delete(ctx context.Context, id int64) error
}

type clubUseCase struct {
	clubRepo repository.ClubRepository
}

func NewClubUseCase(clubRepo repository.ClubRepository) ClubUseCase {
	return &clubUseCase{
		clubRepo: clubRepo,
	}
}

func (uc *clubUseCase) Create(ctx context.Context, club *entity.Club) (int64, error) {
	return uc.clubRepo.Create(ctx, club)
}

func (uc *clubUseCase) GetByID(ctx context.Context, id int64) (*entity.Club, error) {
	return uc.clubRepo.GetByID(ctx, id)
}

func (uc *clubUseCase) GetByTelegramChatID(ctx context.Context, chatID int64) (*entity.Club, error) {
	return uc.clubRepo.GetByTelegramChatID(ctx, chatID)
}

func (uc *clubUseCase) List(ctx context.Context) ([]entity.Club, error) {
	return uc.clubRepo.List(ctx)
}

func (uc *clubUseCase) Update(ctx context.Context, club *entity.Club) error {
	return uc.clubRepo.Update(ctx, club)
}

func (uc *clubUseCase) Delete(ctx context.Context, id int64) error {
	return uc.clubRepo.Delete(ctx, id)
}
