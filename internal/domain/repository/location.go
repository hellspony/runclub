package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type LocationRepository interface {
	Create(ctx context.Context, location *entity.Location) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Location, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Location, error)
	Update(ctx context.Context, location *entity.Location) error
	Delete(ctx context.Context, id int64) error
}
