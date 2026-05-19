package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type ClubRepository interface {
	Create(ctx context.Context, club *entity.Club) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Club, error)
	GetByTelegramChatID(ctx context.Context, chatID int64) (*entity.Club, error)
	List(ctx context.Context) ([]entity.Club, error)
	Update(ctx context.Context, club *entity.Club) error
	Delete(ctx context.Context, id int64) error
}
